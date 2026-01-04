package service

import (
	"fmt"

	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/database/backend"
	"github.com/multimarket-labs/event-pod-services/services/api/models"
)

// CreateEvent creates a new prediction event with sub-events, outcomes, and tags
// All operations are wrapped in a database transaction
func (h *HandlerSvc) CreateEvent(req *models.CreateEventRequest) (*models.CreateEventResponse, error) {
	// Validate request
	if err := h.validateCreateEventRequest(req); err != nil {
		return nil, err
	}

	var response *models.CreateEventResponse
	var eventDB = backend.NewEventDB()

	// Execute all operations in a transaction
	err := h.db.Transaction(func(txDB *database.DB) error {
		now := backend.CurrentTimestamp()
		eventGUID := backend.GenerateCompactUUID()

		// Step 1: Create Event
		event := &backend.Event{
			GUID:        eventGUID,
			Title:       req.Title,
			Description: req.Description,
			ImageURL:    req.ImageURL,
			StartDate:   req.StartDate,
			EndDate:     req.EndDate,
			Created:     now,
			Updated:     now,
		}
		if err := eventDB.CreateEvent(txDB.GetGorm(), event); err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		// Step 2: Process Tags
		var tagResponses []models.TagResponse
		for _, tagName := range req.Tags {
			tag, err := eventDB.GetOrCreateTag(txDB.GetGorm(), tagName)
			if err != nil {
				return fmt.Errorf("failed to get or create tag '%s': %w", tagName, err)
			}

			// Create event-tag association
			eventTag := &backend.EventTag{
				EventGUID: eventGUID,
				TagGUID:   tag.GUID,
				Created:   now,
			}
			if err := eventDB.CreateEventTag(txDB.GetGorm(), eventTag); err != nil {
				return fmt.Errorf("failed to create event-tag association: %w", err)
			}

			tagResponses = append(tagResponses, models.TagResponse{
				GUID:    tag.GUID,
				Name:    tag.Name,
				Created: tag.Created,
				Updated: tag.Updated,
			})
		}

		// Step 3: Create Sub-Events and Outcomes
		var subEventResponses []models.SubEventResponse
		for _, subEventReq := range req.SubEvents {
			subEventGUID := backend.GenerateCompactUUID()

			// Create sub-event
			subEvent := &backend.SubEvent{
				GUID:      subEventGUID,
				EventGUID: eventGUID,
				Question:  subEventReq.Question,
				Created:   now,
				Updated:   now,
			}
			if err := eventDB.CreateSubEvent(txDB.GetGorm(), subEvent); err != nil {
				return fmt.Errorf("failed to create sub-event: %w", err)
			}

			// Step 4: Create Outcomes for this sub-event
			var outcomeResponses []models.OutcomeResponse
			for _, outcomeReq := range subEventReq.Outcomes {
				outcomeGUID := backend.GenerateCompactUUID()

				outcome := &backend.Outcome{
					GUID:         outcomeGUID,
					SubEventGUID: subEventGUID,
					Name:         outcomeReq.Name,
					Color:        outcomeReq.Color,
					Idx:          outcomeReq.Idx,
					Created:      now,
					Updated:      now,
				}
				if err := eventDB.CreateOutcome(txDB.GetGorm(), outcome); err != nil {
					return fmt.Errorf("failed to create outcome: %w", err)
				}

				outcomeResponses = append(outcomeResponses, models.OutcomeResponse{
					GUID:         outcome.GUID,
					SubEventGUID: outcome.SubEventGUID,
					Name:         outcome.Name,
					Color:        outcome.Color,
					Idx:          outcome.Idx,
					Created:      outcome.Created,
					Updated:      outcome.Updated,
				})
			}

			subEventResponses = append(subEventResponses, models.SubEventResponse{
				GUID:      subEvent.GUID,
				EventGUID: subEvent.EventGUID,
				Question:  subEvent.Question,
				Outcomes:  outcomeResponses,
				Created:   subEvent.Created,
				Updated:   subEvent.Updated,
			})
		}

		// Build response
		response = &models.CreateEventResponse{
			GUID:        event.GUID,
			Title:       event.Title,
			Description: event.Description,
			ImageURL:    event.ImageURL,
			StartDate:   event.StartDate,
			EndDate:     event.EndDate,
			Tags:        tagResponses,
			SubEvents:   subEventResponses,
			Created:     event.Created,
			Updated:     event.Updated,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// validateCreateEventRequest validates the create event request
func (h *HandlerSvc) validateCreateEventRequest(req *models.CreateEventRequest) error {
	// Title cannot be empty
	if req.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Must have at least one sub-event
	if len(req.SubEvents) == 0 {
		return fmt.Errorf("at least one sub-event is required")
	}

	// Each sub-event must have at least two outcomes
	for i, subEvent := range req.SubEvents {
		if len(subEvent.Outcomes) < 2 {
			return fmt.Errorf("sub-event %d must have at least 2 outcomes", i)
		}
	}

	// Start date must be before end date
	if req.StartDate >= req.EndDate {
		return fmt.Errorf("start_date must be before end_date")
	}

	return nil
}
