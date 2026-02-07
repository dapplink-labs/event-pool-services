package database

import (
	"fmt"

	"gorm.io/gorm"
)

// EventRepository 事件数据库操作接口
type EventRepository interface {
	// CreateEvent 创建事件
	CreateEvent(db *gorm.DB, event *Event) error
	// CreateEventLanguage 创建事件多语言
	CreateEventLanguage(db *gorm.DB, eventLang *EventLanguage) error
	// CreateSubEvent 创建子事件
	CreateSubEvent(db *gorm.DB, subEvent *SubEvent) error
	// CreateSubEventDirection 创建子事件方向
	CreateSubEventDirection(db *gorm.DB, direction *SubEventDirection) error
	// ListEvents 查询事件列表（支持多语言）
	ListEvents(db *gorm.DB, languageGUID, categoryGUID string, isLive *int16, page, limit int) ([]Event, int64, error)
	// GetEventWithLanguage 获取事件及其多语言信息
	GetEventWithLanguage(db *gorm.DB, eventGUID, languageGUID string) (*Event, *EventLanguage, error)
	// GetSubEventsByEventGUID 获取事件的所有子事件
	GetSubEventsByEventGUID(db *gorm.DB, eventGUID string) ([]SubEvent, error)
	// GetSubEventDirections 获取子事件的所有方向
	GetSubEventDirections(db *gorm.DB, subEventGUID string) ([]SubEventDirection, error)
}

type eventRepository struct{}

// NewEventRepository 创建事件仓储实例
func NewEventRepository() EventRepository {
	return &eventRepository{}
}

func (r *eventRepository) CreateEvent(db *gorm.DB, event *Event) error {
	return db.Create(event).Error
}

func (r *eventRepository) CreateEventLanguage(db *gorm.DB, eventLang *EventLanguage) error {
	return db.Create(eventLang).Error
}

func (r *eventRepository) CreateSubEvent(db *gorm.DB, subEvent *SubEvent) error {
	return db.Create(subEvent).Error
}

func (r *eventRepository) CreateSubEventDirection(db *gorm.DB, direction *SubEventDirection) error {
	return db.Create(direction).Error
}

// ListEvents 查询事件列表（支持多语言）
func (r *eventRepository) ListEvents(db *gorm.DB, languageGUID, categoryGUID string, isLive *int16, page, limit int) ([]Event, int64, error) {
	var events []Event
	var total int64

	// 构建查询
	query := db.Model(&Event{})

	// 按分类过滤
	if categoryGUID != "" {
		query = query.Where("category_guid = ?", categoryGUID)
	}

	// 按状态过滤
	if isLive != nil {
		query = query.Where("is_live = ?", *isLive)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}

	return events, total, nil
}

// GetEventWithLanguage 获取事件及其多语言信息
func (r *eventRepository) GetEventWithLanguage(db *gorm.DB, eventGUID, languageGUID string) (*Event, *EventLanguage, error) {
	var event Event
	if err := db.Where("guid = ?", eventGUID).First(&event).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get event: %w", err)
	}

	var eventLang EventLanguage
	if err := db.Where("event_guid = ? AND language_guid = ?", eventGUID, languageGUID).First(&eventLang).Error; err != nil {
		return &event, nil, fmt.Errorf("failed to get event language: %w", err)
	}

	return &event, &eventLang, nil
}

// GetSubEventsByEventGUID 获取事件的所有子事件
func (r *eventRepository) GetSubEventsByEventGUID(db *gorm.DB, eventGUID string) ([]SubEvent, error) {
	var subEvents []SubEvent
	if err := db.Where("parent_event_guid = ?", eventGUID).Order("created_at ASC").Find(&subEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get sub events: %w", err)
	}
	return subEvents, nil
}

// GetSubEventDirections 获取子事件的所有方向
func (r *eventRepository) GetSubEventDirections(db *gorm.DB, subEventGUID string) ([]SubEventDirection, error) {
	var directions []SubEventDirection
	if err := db.Where("sub_event_guid = ?", subEventGUID).Order("created_at ASC").Find(&directions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sub event directions: %w", err)
	}
	return directions, nil
}
