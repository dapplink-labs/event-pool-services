package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// DifyResponse represents the first-layer response from Dify API
type DifyResponse struct {
	Data struct {
		Outputs struct {
			EventData string `json:"event_data"` // Escaped JSON string that needs second parsing
		} `json:"outputs"`
		Status string `json:"status"`
	} `json:"data"`
	WorkflowRunID string `json:"workflow_run_id"`
	TaskID        string `json:"task_id"`
}

// EventDetail represents the actual prediction event content (second-layer parsing)
type EventDetail struct {
	EventTitle   string   `json:"event_title"`
	Content      string   `json:"content"`
	StartAt      int64    `json:"start_at"`
	EndAt        int64    `json:"end_at"`
	CategoryTags []string `json:"category_tags"`
	Questions    []struct {
		Title   string `json:"title"`
		Options []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"options"`
	} `json:"questions"`
}

// DifyRequest represents the request body for Dify workflow
type DifyRequest struct {
	Inputs       map[string]string `json:"inputs"`
	ResponseMode string            `json:"response_mode"`
	User         string            `json:"user"`
}

const (
	difyAPIURL     = "https://api.dify.ai/v1/workflows/run"
	requestTimeout = 120 * time.Second // 增加到 120 秒
	maxRetries     = 2                 // 最大重试次数
)

// GetPredictEvent calls Dify workflow API to convert natural language query into structured event data
func (s *HandlerSvc) GetPredictEvent(ctx context.Context, userQuery string) (*EventDetail, error) {
	log.Printf("[Dify] Starting prediction for query: %s", userQuery)
	var lastErr error

	// Retry logic
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.Printf("[Dify] Retry attempt %d/%d, waiting %v...", attempt, maxRetries, waitTime)
			time.Sleep(waitTime)
		}

		log.Printf("[Dify] Attempt %d: Calling Dify API...", attempt+1)
		eventDetail, err := s.callDifyAPI(ctx, userQuery)
		if err == nil {
			log.Printf("[Dify] Success on attempt %d", attempt+1)
			return eventDetail, nil
		}

		log.Printf("[Dify] Attempt %d failed: %v", attempt+1, err)
		lastErr = err

		// Don't retry on validation errors
		if isValidationError(err) {
			log.Printf("[Dify] Validation error detected, not retrying")
			return nil, err
		}
	}

	log.Printf("[Dify] All attempts failed after %d retries", maxRetries)
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// callDifyAPI performs the actual API call
func (s *HandlerSvc) callDifyAPI(ctx context.Context, userQuery string) (*EventDetail, error) {
	startTime := time.Now()

	// Get API key from environment variable
	apiKey := os.Getenv("DIFY_API_KEY")
	if apiKey == "" {
		log.Printf("[Dify] ERROR: DIFY_API_KEY not set")
		return nil, fmt.Errorf("DIFY_API_KEY environment variable is not set")
	}

	// Get current date for sys_date parameter
	sysDate := time.Now().Format("2006-01-02")

	// Construct request body
	reqBody := DifyRequest{
		Inputs: map[string]string{
			"user_query": userQuery,
			"sys_date":   sysDate,
		},
		ResponseMode: "blocking",
		User:         "system_integration",
	}

	// Marshal request body to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[Dify] ERROR: Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Log request info
	log.Printf("[Dify] Request: POST %s, Query: %s", difyAPIURL, userQuery)

	// Create HTTP client with timeout
	// Use system proxy if available
	var transport *http.Transport

	// Try to use proxy from environment
	proxyURL := os.Getenv("HTTPS_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTP_PROXY")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("https_proxy")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("http_proxy")
	}

	if proxyURL != "" {
		log.Printf("[Dify] Using proxy: %s", proxyURL)
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
	} else {
		transport = &http.Transport{}
	}

	client := &http.Client{
		Timeout:   requestTimeout,
		Transport: transport,
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", difyAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Dify] ERROR: Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		elapsed := time.Since(startTime)
		log.Printf("[Dify] ERROR: Request failed after %v: %v", elapsed, err)
		return nil, fmt.Errorf("failed to send request to Dify API: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Dify] ERROR: Failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("[Dify] ERROR: Status %d, Response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Dify API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// First-layer parsing: Parse Dify response wrapper
	var difyResp DifyResponse
	if err := json.Unmarshal(body, &difyResp); err != nil {
		log.Printf("[Dify] ERROR: Failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse Dify response (first layer): %w, body: %s", err, string(body))
	}

	// Check if status is successful
	if difyResp.Data.Status != "succeeded" {
		log.Printf("[Dify] ERROR: Workflow status: %s", difyResp.Data.Status)
		return nil, fmt.Errorf("Dify workflow status is not succeeded: %s", difyResp.Data.Status)
	}

	// Check if event_data is not empty
	if difyResp.Data.Outputs.EventData == "" {
		log.Printf("[Dify] ERROR: event_data is empty")
		return nil, fmt.Errorf("event_data field is empty in Dify response")
	}

	// Second-layer parsing: Parse the escaped JSON string into EventDetail
	var eventDetail EventDetail
	if err := json.Unmarshal([]byte(difyResp.Data.Outputs.EventData), &eventDetail); err != nil {
		log.Printf("[Dify] ERROR: Failed to parse event_data: %v", err)
		return nil, fmt.Errorf("failed to parse event_data (second layer): %w, event_data: %s", err, difyResp.Data.Outputs.EventData)
	}

	// Validate required fields
	if eventDetail.EventTitle == "" {
		log.Printf("[Dify] ERROR: event_title is empty")
		return nil, fmt.Errorf("event_title is empty in parsed event data")
	}
	if eventDetail.StartAt == 0 || eventDetail.EndAt == 0 {
		log.Printf("[Dify] ERROR: start_at or end_at is missing")
		return nil, fmt.Errorf("start_at or end_at is missing in parsed event data")
	}
	if len(eventDetail.Questions) == 0 {
		log.Printf("[Dify] ERROR: questions array is empty")
		return nil, fmt.Errorf("questions array is empty in parsed event data")
	}

	// Success log
	log.Printf("[Dify] SUCCESS: Generated event '%s' with %d questions in %v", eventDetail.EventTitle, len(eventDetail.Questions), elapsed)
	return &eventDetail, nil
}

// isValidationError checks if the error is a validation error that should not be retried
func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "DIFY_API_KEY") ||
		contains(errMsg, "event_title is empty") ||
		contains(errMsg, "questions array is empty")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
