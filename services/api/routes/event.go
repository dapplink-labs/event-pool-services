package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/log"

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
