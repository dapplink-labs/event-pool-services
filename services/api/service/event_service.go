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

// ListEvents retrieves events with filtering and pagination
func (h *HandlerSvc) ListEvents(req *models.ListEventsRequest) (*models.ListEventsResponse, error) {
	// Validate and normalize pagination parameters
	page, limit := validatePagination(req.Page, req.Limit)

	var eventDB = backend.NewEventDB()

	// Retrieve events from database (status filter will be applied after retrieval)
	events, total, err := eventDB.ListEvents(
		h.db.GetGorm(),
		req.Tags,
		"", // status will be calculated
		req.Keyword,
		req.StartTime,
		req.EndTime,
		page,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Build response
	var eventItems []models.EventListItem
	for _, event := range events {
		// Calculate status
		status := calculateStatus(event.StartDate, event.EndDate)

		// Skip if status filter doesn't match
		if req.Status != "" && status != req.Status {
			continue
		}

		// Get tags for this event
		tags, err := eventDB.GetEventTags(h.db.GetGorm(), event.GUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for event %s: %w", event.GUID, err)
		}

		// Extract tag names
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}

		eventItems = append(eventItems, models.EventListItem{
			GUID:        event.GUID,
			Title:       event.Title,
			Description: event.Description,
			ImageURL:    event.ImageURL,
			StartDate:   event.StartDate,
			EndDate:     event.EndDate,
			Status:      status,
			Tags:        tagNames,
			Created:     event.Created,
			Updated:     event.Updated,
		})
	}

	// Calculate pagination info
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := &models.ListEventsResponse{
		Events: eventItems,
		Pagination: models.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	return response, nil
}

// GetEventDetail retrieves full event details including sub-events and outcomes
func (h *HandlerSvc) GetEventDetail(guid string) (*models.GetEventDetailResponse, error) {
	var eventDB = backend.NewEventDB()

	// Get event
	event, err := eventDB.GetEventByGUID(h.db.GetGorm(), guid)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Calculate status
	status := calculateStatus(event.StartDate, event.EndDate)

	// Get tags
	tags, err := eventDB.GetEventTags(h.db.GetGorm(), event.GUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	// Extract tag names
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	// Get sub-events
	subEvents, err := eventDB.GetEventSubEvents(h.db.GetGorm(), event.GUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-events: %w", err)
	}

	// Build sub-events response with outcomes
	var subEventResponses []models.SubEventResponse
	for _, subEvent := range subEvents {
		// Get outcomes for this sub-event
		outcomes, err := eventDB.GetSubEventOutcomes(h.db.GetGorm(), subEvent.GUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get outcomes for sub-event %s: %w", subEvent.GUID, err)
		}

		// Build outcomes response
		var outcomeResponses []models.OutcomeResponse
		for _, outcome := range outcomes {
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
	response := &models.GetEventDetailResponse{
		GUID:        event.GUID,
		Title:       event.Title,
		Description: event.Description,
		ImageURL:    event.ImageURL,
		StartDate:   event.StartDate,
		EndDate:     event.EndDate,
		Status:      status,
		Tags:        tagNames,
		SubEvents:   subEventResponses,
		Created:     event.Created,
		Updated:     event.Updated,
	}

	return response, nil
}

// calculateStatus determines event status based on current time
func calculateStatus(startDate, endDate int64) string {
	now := backend.CurrentTimestamp()
	if now < startDate {
		return "upcoming"
	} else if now >= startDate && now <= endDate {
		return "active"
	}
	return "ended"
}

// validatePagination validates and normalizes pagination parameters
func validatePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
