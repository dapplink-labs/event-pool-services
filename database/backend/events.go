package backend

import "gorm.io/gorm"

// Event 预测事件主表
type Event struct {
	GUID        string `gorm:"type:varchar(32);primaryKey;comment:事件唯一标识" json:"guid"`
	Title       string `gorm:"type:varchar(255);not null;comment:事件标题" json:"title"`
	Description string `gorm:"type:text;comment:事件描述" json:"description"`
	ImageURL    string `gorm:"type:varchar(512);comment:事件图片URL" json:"image_url"`
	StartDate   int64  `gorm:"type:bigint;not null;comment:开始时间(Unix时间戳)" json:"start_date"`
	EndDate     int64  `gorm:"type:bigint;not null;comment:结束时间(Unix时间戳)" json:"end_date"`
	Created     int64  `gorm:"type:bigint;not null;comment:创建时间(Unix时间戳)" json:"created"`
	Updated     int64  `gorm:"type:bigint;not null;comment:更新时间(Unix时间戳)" json:"updated"`
}

// TableName specifies the table name for Event model
func (Event) TableName() string {
	return "events"
}

// SubEvent 子事件表（问题）
type SubEvent struct {
	GUID      string `gorm:"type:varchar(32);primaryKey;comment:子事件唯一标识" json:"guid"`
	EventGUID string `gorm:"type:varchar(32);not null;index;comment:关联事件GUID" json:"event_guid"`
	Question  string `gorm:"type:text;not null;comment:问题内容" json:"question"`
	Created   int64  `gorm:"type:bigint;not null;comment:创建时间(Unix时间戳)" json:"created"`
	Updated   int64  `gorm:"type:bigint;not null;comment:更新时间(Unix时间戳)" json:"updated"`
}

// TableName specifies the table name for SubEvent model
func (SubEvent) TableName() string {
	return "sub_events"
}

// Outcome 结果选项表
type Outcome struct {
	GUID         string `gorm:"type:varchar(32);primaryKey;comment:结果唯一标识" json:"guid"`
	SubEventGUID string `gorm:"type:varchar(32);not null;index;comment:关联子事件GUID" json:"sub_event_guid"`
	Name         string `gorm:"type:varchar(255);not null;comment:结果名称" json:"name"`
	Color        string `gorm:"type:varchar(20);comment:结果颜色" json:"color"`
	Idx          int    `gorm:"type:int;not null;comment:排序索引" json:"idx"`
	Created      int64  `gorm:"type:bigint;not null;comment:创建时间(Unix时间戳)" json:"created"`
	Updated      int64  `gorm:"type:bigint;not null;comment:更新时间(Unix时间戳)" json:"updated"`
}

// TableName specifies the table name for Outcome model
func (Outcome) TableName() string {
	return "outcomes"
}

// Tag 标签表
type Tag struct {
	GUID    string `gorm:"type:varchar(32);primaryKey;comment:标签唯一标识" json:"guid"`
	Name    string `gorm:"type:varchar(100);not null;uniqueIndex;comment:标签名称" json:"name"`
	Created int64  `gorm:"type:bigint;not null;comment:创建时间(Unix时间戳)" json:"created"`
	Updated int64  `gorm:"type:bigint;not null;comment:更新时间(Unix时间戳)" json:"updated"`
}

// TableName specifies the table name for Tag model
func (Tag) TableName() string {
	return "tags"
}

// EventTag 事件-标签关联表
type EventTag struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	EventGUID string `gorm:"type:varchar(32);not null;index;comment:事件GUID" json:"event_guid"`
	TagGUID   string `gorm:"type:varchar(32);not null;index;comment:标签GUID" json:"tag_guid"`
	Created   int64  `gorm:"type:bigint;not null;comment:创建时间(Unix时间戳)" json:"created"`
}

// TableName specifies the table name for EventTag model
func (EventTag) TableName() string {
	return "event_tags"
}

// EventDB 数据库接口
type EventDB interface {
	CreateEvent(db *gorm.DB, event *Event) error
	CreateSubEvent(db *gorm.DB, subEvent *SubEvent) error
	CreateOutcome(db *gorm.DB, outcome *Outcome) error
	GetOrCreateTag(db *gorm.DB, tagName string) (*Tag, error)
	CreateEventTag(db *gorm.DB, eventTag *EventTag) error
	ListEvents(db *gorm.DB, tags []string, status string, keyword string, startTime, endTime int64, page, limit int) ([]Event, int64, error)
	GetEventByGUID(db *gorm.DB, guid string) (*Event, error)
	GetEventTags(db *gorm.DB, eventGUID string) ([]Tag, error)
	GetEventSubEvents(db *gorm.DB, eventGUID string) ([]SubEvent, error)
	GetSubEventOutcomes(db *gorm.DB, subEventGUID string) ([]Outcome, error)
}

type eventDB struct{}

// NewEventDB creates a new EventDB instance
func NewEventDB() EventDB {
	return &eventDB{}
}

func (edb *eventDB) CreateEvent(db *gorm.DB, event *Event) error {
	return db.Create(event).Error
}

func (edb *eventDB) CreateSubEvent(db *gorm.DB, subEvent *SubEvent) error {
	return db.Create(subEvent).Error
}

func (edb *eventDB) CreateOutcome(db *gorm.DB, outcome *Outcome) error {
	return db.Create(outcome).Error
}

func (edb *eventDB) GetOrCreateTag(db *gorm.DB, tagName string) (*Tag, error) {
	var tag Tag
	err := db.Where("name = ?", tagName).First(&tag).Error
	if err == nil {
		// Tag already exists
		return &tag, nil
	}
	if err != gorm.ErrRecordNotFound {
		// Database error
		return nil, err
	}

	// Create new tag
	tag = Tag{
		GUID:    GenerateCompactUUID(),
		Name:    tagName,
		Created: CurrentTimestamp(),
		Updated: CurrentTimestamp(),
	}
	if err := db.Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (edb *eventDB) CreateEventTag(db *gorm.DB, eventTag *EventTag) error {
	return db.Create(eventTag).Error
}

// ListEvents retrieves events with filtering and pagination
func (edb *eventDB) ListEvents(db *gorm.DB, tags []string, status string, keyword string, startTime, endTime int64, page, limit int) ([]Event, int64, error) {
	var events []Event
	var total int64

	// Build query
	query := db.Model(&Event{})

	// Filter by tags (if provided)
	if len(tags) > 0 {
		query = query.Joins("JOIN event_tags ON event_tags.event_guid = events.guid").
			Joins("JOIN tags ON tags.guid = event_tags.tag_guid").
			Where("tags.name IN ?", tags).
			Group("events.guid").
			Having("COUNT(DISTINCT tags.name) = ?", len(tags))
	}

	// Filter by keyword (search in title and description)
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Filter by time range
	if startTime > 0 {
		query = query.Where("start_date >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("end_date <= ?", endTime)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	offset := (page - 1) * limit
	if err := query.Order("created DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventByGUID retrieves a single event by GUID
func (edb *eventDB) GetEventByGUID(db *gorm.DB, guid string) (*Event, error) {
	var event Event
	if err := db.Where("guid = ?", guid).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// GetEventTags retrieves all tags for a specific event
func (edb *eventDB) GetEventTags(db *gorm.DB, eventGUID string) ([]Tag, error) {
	var tags []Tag
	err := db.Joins("JOIN event_tags ON event_tags.tag_guid = tags.guid").
		Where("event_tags.event_guid = ?", eventGUID).
		Find(&tags).Error
	return tags, err
}

// GetEventSubEvents retrieves all sub-events for a specific event
func (edb *eventDB) GetEventSubEvents(db *gorm.DB, eventGUID string) ([]SubEvent, error) {
	var subEvents []SubEvent
	err := db.Where("event_guid = ?", eventGUID).Order("created ASC").Find(&subEvents).Error
	return subEvents, err
}

// GetSubEventOutcomes retrieves all outcomes for a specific sub-event
func (edb *eventDB) GetSubEventOutcomes(db *gorm.DB, subEventGUID string) ([]Outcome, error) {
	var outcomes []Outcome
	err := db.Where("sub_event_guid = ?", subEventGUID).Order("idx ASC").Find(&outcomes).Error
	return outcomes, err
}
