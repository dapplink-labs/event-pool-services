package service

import (
	"fmt"
	"time"

	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/services/api/models"
)

// CreateEvent 创建新的预测事件（基于新表结构）
// 逻辑流程：
// 1. 开启事务
// 2. 插入 event 表
// 3. 插入 event_language 表
// 4. 插入 sub_event 表
// 5. 插入 sub_event_direction 表
// 6. 提交事务
func (h *HandlerSvc) CreateEvent(req *models.CreateEventRequest) (*models.CreateEventResponse, error) {
	// 验证请求
	if err := h.validateCreateEventNewRequest(req); err != nil {
		return nil, err
	}

	var response *models.CreateEventResponse
	repo := database.NewEventRepository()

	// 在事务中执行所有操作
	err := h.db.Transaction(func(txDB *database.DB) error {
		db := txDB.GetGorm()

		// Step 1: 创建 Event（GUID 由数据库自动生成）
		event := &database.Event{
			CategoryGUID:         req.CategoryGUID,
			EcosystemGUID:        req.EcosystemGUID,
			EventPeriodGUID:      req.EventPeriodGUID,
			MainTeamGroupGUID:    req.MainTeamGroupGUID,
			ClusterTeamGroupGUID: req.ClusterTeamGroupGUID,
			MainScore:            "0", // 初始分数为 0
			ClusterScore:         "0", // 初始分数为 0
			Logo:                 req.Logo,
			OrderType:            0,   // 默认为热门话题
			OrderNum:             "0", // 初始订单数为 0
			OpenTime:             "",  // 开盘时间稍后设置
			TradeVolume:          0,   // 初始交易量为 0
			ExperimentResult:     "",  // 实验结果为空
			Info:                 database.JSONB{},
			IsOnline:             false, // 默认不上线
			IsLive:               1,     // 默认为未来事件
			IsSports:             req.IsSports,
			Stage:                "Q1", // 默认阶段
		}

		if err := repo.CreateEvent(db, event); err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		// 获取数据库生成的 GUID
		if err := db.Where("category_guid = ? AND ecosystem_guid = ? AND event_period_guid = ?",
			req.CategoryGUID, req.EcosystemGUID, req.EventPeriodGUID).
			Order("created_at DESC").First(event).Error; err != nil {
			return fmt.Errorf("failed to retrieve created event: %w", err)
		}

		eventGUID := event.GUID

		// Step 2: 创建 EventLanguage
		eventLang := &database.EventLanguage{
			EventGUID:    eventGUID,
			LanguageGUID: req.LanguageGUID,
			Title:        req.Title,
			Rules:        req.Rules,
		}

		if err := repo.CreateEventLanguage(db, eventLang); err != nil {
			return fmt.Errorf("failed to create event language: %w", err)
		}

		// Step 3 & 4: 创建 SubEvent 和 SubEventDirection
		var subEventResponses []models.SubEventResponse
		for _, subEventReq := range req.SubEvents {
			// 创建子事件
			subEvent := &database.SubEvent{
				ParentEventGUID: eventGUID,
				Title:           subEventReq.Title,
				Logo:            req.Logo, // 使用事件的 Logo
				TradeVolume:     0,
			}

			if err := repo.CreateSubEvent(db, subEvent); err != nil {
				return fmt.Errorf("failed to create sub event: %w", err)
			}

			// 获取数据库生成的 GUID
			if err := db.Where("parent_event_guid = ? AND title = ?", eventGUID, subEventReq.Title).
				Order("created_at DESC").First(subEvent).Error; err != nil {
				return fmt.Errorf("failed to retrieve created sub event: %w", err)
			}

			subEventGUID := subEvent.GUID

			// 创建子事件方向
			var directionResponses []models.SubEventDirectionResponse
			for _, dirReq := range subEventReq.Directions {
				direction := &database.SubEventDirection{
					SubEventGUID: subEventGUID,
					Direction:    dirReq.Direction,
					Chance:       dirReq.Chance,
					NewAskPrice:  "0",
					NewBidPrice:  "0",
					Info:         database.JSONB{},
				}

				if err := repo.CreateSubEventDirection(db, direction); err != nil {
					return fmt.Errorf("failed to create sub event direction: %w", err)
				}

				// 获取数据库生成的 GUID
				if err := db.Where("sub_event_guid = ? AND direction = ?", subEventGUID, dirReq.Direction).
					Order("created_at DESC").First(direction).Error; err != nil {
					return fmt.Errorf("failed to retrieve created direction: %w", err)
				}

				directionResponses = append(directionResponses, models.SubEventDirectionResponse{
					GUID:        direction.GUID,
					Direction:   direction.Direction,
					Chance:      direction.Chance,
					NewAskPrice: direction.NewAskPrice,
					NewBidPrice: direction.NewBidPrice,
				})
			}

			subEventResponses = append(subEventResponses, models.SubEventResponse{
				GUID:       subEvent.GUID,
				Title:      subEvent.Title,
				Logo:       subEvent.Logo,
				Directions: directionResponses,
			})
		}

		// 构建响应
		response = &models.CreateEventResponse{
			GUID:      eventGUID,
			Title:     req.Title,
			Rules:     req.Rules,
			Logo:      req.Logo,
			SubEvents: subEventResponses,
			CreatedAt: event.CreatedAt.Format(time.RFC3339),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// ListEvents 查询事件列表（基于新表结构，支持多语言）
func (h *HandlerSvc) ListEvents(req *models.ListEventsRequest) (*models.ListEventsResponse, error) {
	// 验证请求
	if req.LanguageGUID == "" {
		return nil, fmt.Errorf("language_guid is required")
	}

	// 验证和规范化分页参数
	page, limit := validatePagination(req.Page, req.Limit)

	repo := database.NewEventRepository()
	db := h.db.GetGorm()

	// 查询事件列表
	events, total, err := repo.ListEvents(db, req.LanguageGUID, req.CategoryGUID, req.IsLive, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// 构建响应
	var eventItems []models.EventListItem
	for _, event := range events {
		// 获取事件的多语言信息
		var eventLang database.EventLanguage
		if err := db.Where("event_guid = ? AND language_guid = ?", event.GUID, req.LanguageGUID).
			First(&eventLang).Error; err != nil {
			// 如果没有找到对应语言，跳过该事件
			continue
		}

		// 获取子事件列表
		subEvents, err := repo.GetSubEventsByEventGUID(db, event.GUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sub events for event %s: %w", event.GUID, err)
		}

		// 构建子事件响应
		var subEventResponses []models.SubEventResponse
		for _, subEvent := range subEvents {
			// 获取子事件方向
			directions, err := repo.GetSubEventDirections(db, subEvent.GUID)
			if err != nil {
				return nil, fmt.Errorf("failed to get directions for sub event %s: %w", subEvent.GUID, err)
			}

			// 构建方向响应
			var directionResponses []models.SubEventDirectionResponse
			for _, dir := range directions {
				directionResponses = append(directionResponses, models.SubEventDirectionResponse{
					GUID:        dir.GUID,
					Direction:   dir.Direction,
					Chance:      dir.Chance,
					NewAskPrice: dir.NewAskPrice,
					NewBidPrice: dir.NewBidPrice,
				})
			}

			subEventResponses = append(subEventResponses, models.SubEventResponse{
				GUID:       subEvent.GUID,
				Title:      subEvent.Title,
				Logo:       subEvent.Logo,
				Directions: directionResponses,
			})
		}

		eventItems = append(eventItems, models.EventListItem{
			GUID:            event.GUID,
			Title:           eventLang.Title,
			Rules:           eventLang.Rules,
			Logo:            event.Logo,
			CategoryGUID:    event.CategoryGUID,
			EcosystemGUID:   event.EcosystemGUID,
			EventPeriodGUID: event.EventPeriodGUID,
			IsLive:          event.IsLive,
			IsSports:        event.IsSports,
			OpenTime:        event.OpenTime,
			TradeVolume:     event.TradeVolume,
			SubEvents:       subEventResponses,
			CreatedAt:       event.CreatedAt.Format(time.RFC3339),
		})
	}

	// 计算分页信息
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

// validateCreateEventNewRequest 验证创建事件请求
func (h *HandlerSvc) validateCreateEventNewRequest(req *models.CreateEventRequest) error {
	if req.CategoryGUID == "" {
		return fmt.Errorf("category_guid is required")
	}
	if req.EcosystemGUID == "" {
		return fmt.Errorf("ecosystem_guid is required")
	}
	if req.EventPeriodGUID == "" {
		return fmt.Errorf("event_period_guid is required")
	}
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.LanguageGUID == "" {
		return fmt.Errorf("language_guid is required")
	}
	if len(req.SubEvents) == 0 {
		return fmt.Errorf("at least one sub_event is required")
	}

	// 验证每个子事件至少有两个方向
	for i, subEvent := range req.SubEvents {
		if len(subEvent.Directions) < 2 {
			return fmt.Errorf("sub_event %d must have at least 2 directions", i)
		}
	}

	return nil
}

// validatePagination 验证和规范化分页参数
func validatePagination(page, limit int) (int, int) {
	// 默认页码为 1
	if page < 1 {
		page = 1
	}

	// 默认每页 20 条，最大 100 条
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}
