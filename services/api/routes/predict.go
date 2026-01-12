package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/multimarket-labs/event-pod-services/services/api/models"
)

// PredictEventRequest represents the request body for predict event endpoint
type PredictEventRequest struct {
	UserQuery string `json:"user_query" binding:"required"` // Natural language query from user
}

// PredictEventHandler handles POST /api/v1/admin/predict-event
// This endpoint calls Dify workflow to convert natural language into structured event data
func (rs *Routes) PredictEventHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req PredictEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserQuery == "" {
		jsonResponse(w, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "user_query is required",
		}, http.StatusBadRequest)
		return
	}

	// Create a new context with longer timeout for Dify API call
	// Dify workflow can take a long time (LLM inference)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Call Dify service to get predicted event
	eventDetail, err := rs.svc.GetPredictEvent(ctx, req.UserQuery)
	if err != nil {
		jsonResponse(w, models.ErrorResponse{
			Error:   "prediction_failed",
			Message: "Failed to predict event: " + err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Return the predicted event detail
	jsonResponse(w, eventDetail, http.StatusOK)
}
