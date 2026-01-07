package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/go-chi/chi/v5"

	"github.com/multimarket-labs/event-pod-services/services/api/models"
)

// CreateEventHandler handles POST /api/v1/admin/events
func (rs *Routes) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request body", "err", err)
		jsonResponse(w, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		}, http.StatusBadRequest)
		return
	}

	// Call service layer
	response, err := rs.svc.CreateEvent(&req)
	if err != nil {
		log.Error("failed to create event", "err", err)
		jsonResponse(w, models.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Return success response
	jsonResponse(w, response, http.StatusCreated)
}

// ListEventsHandler handles GET /api/v1/admin/events
func (rs *Routes) ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := models.ListEventsRequest{
		Page:  1,
		Limit: 20,
	}

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	// Parse tags (comma-separated)
	if tagsStr := r.URL.Query().Get("tags"); tagsStr != "" {
		req.Tags = strings.Split(tagsStr, ",")
		// Trim spaces
		for i := range req.Tags {
			req.Tags[i] = strings.TrimSpace(req.Tags[i])
		}
	}

	// Parse status
	req.Status = r.URL.Query().Get("status")

	// Parse keyword
	req.Keyword = r.URL.Query().Get("keyword")

	// Parse start_time
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			req.StartTime = startTime
		}
	}

	// Parse end_time
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			req.EndTime = endTime
		}
	}

	// Call service layer
	response, err := rs.svc.ListEvents(&req)
	if err != nil {
		log.Error("failed to list events", "err", err)
		jsonResponse(w, models.ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Return success response
	jsonResponse(w, response, http.StatusOK)
}

// GetEventDetailHandler handles GET /api/v1/admin/events/:guid
func (rs *Routes) GetEventDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Get GUID from URL parameter
	guid := chi.URLParam(r, "guid")
	if guid == "" {
		log.Error("event guid is required")
		jsonResponse(w, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Event GUID is required",
		}, http.StatusBadRequest)
		return
	}

	// Call service layer
	response, err := rs.svc.GetEventDetail(guid)
	if err != nil {
		log.Error("failed to get event detail", "err", err, "guid", guid)
		jsonResponse(w, models.ErrorResponse{
			Error:   "get_failed",
			Message: err.Error(),
		}, http.StatusNotFound)
		return
	}

	// Return success response
	jsonResponse(w, response, http.StatusOK)
}
