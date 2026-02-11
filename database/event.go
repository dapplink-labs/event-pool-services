package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type Event struct {
	GUID                 string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	CategoryGUID         string    `gorm:"type:varchar(500);not null" json:"category_guid"`
	EcosystemGUID        string    `gorm:"type:varchar(500);not null" json:"ecosystem_guid"`
	EventPeriodGUID      string    `gorm:"type:varchar(500);not null" json:"event_period_guid"`
	MainTeamGroupGUID    string    `gorm:"type:varchar(255);not null;default:'0'" json:"main_team_group_guid"`
	ClusterTeamGroupGUID string    `gorm:"type:varchar(255);not null;default:'0'" json:"cluster_team_group_guid"`
	ExternalId           string    `gorm:"type:varchar(255);not null;default:'0'" json:"external_id"`
	MainScore            string    `gorm:"type:numeric;not null;default:0" json:"main_score"`    // UINT256 mapped to string
	ClusterScore         string    `gorm:"type:numeric;not null;default:0" json:"cluster_score"` // UINT256 mapped to string
	Logo                 string    `gorm:"type:varchar(500);not null" json:"logo"`
	EventType            int16     `gorm:"type:smallint;not null;default:0;column:event_type" json:"event_type"` // Event type: 0=hot topic, 1=breaking, 2=latest
	ExperimentResult     string    `gorm:"type:text;not null;default:''" json:"experiment_result"`
	Info                 JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"info"`
	IsOnline             bool      `gorm:"type:boolean;not null;default:false" json:"is_online"`
	IsLive               int16     `gorm:"type:smallint;not null;default:0" json:"is_live"`
	IsSports             bool      `gorm:"type:boolean;not null;default:true" json:"is_sports"`
	Stage                string    `gorm:"type:varchar(20);not null;default:'Q1'" json:"stage"`
	CreatedAt            time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Event) TableName() string {
	return "event"
}

type EventLanguage struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	EventGUID    string    `gorm:"type:varchar(500);not null" json:"event_guid"`
	LanguageGUID string    `gorm:"type:varchar(500);not null" json:"language_guid"`
	Title        string    `gorm:"type:varchar(200);not null" json:"title"`
	Rules        string    `gorm:"type:text;not null" json:"rules"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EventLanguage) TableName() string {
	return "event_language"
}

type SubEvent struct {
	GUID            string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	ParentEventGUID string    `gorm:"type:varchar(500);not null;index:idx_sub_event_parent_event_guid" json:"parent_event_guid"`
	Title           string    `gorm:"type:varchar(200);not null" json:"title"`
	Logo            string    `gorm:"type:varchar(300);not null" json:"logo"`
	TradeVolume     float64   `gorm:"type:numeric(32,16);not null;default:0" json:"trade_volume"`
	CreatedAt       time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SubEvent) TableName() string {
	return "sub_event"
}

type SubEventLanguage struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LanguageGUID string    `gorm:"type:varchar(500);not null" json:"language_guid"`
	SubEventGUID string    `gorm:"type:varchar(500);not null" json:"sub_event_guid"`
	Title        string    `gorm:"type:varchar(200);not null" json:"title"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SubEventLanguage) TableName() string {
	return "sub_event_language"
}

type SubEventDirection struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	SubEventGUID string    `gorm:"type:varchar(500);not null;index:idx_sub_event_direction_sub_event_guid" json:"sub_event_guid"`
	Direction    string    `gorm:"type:varchar(200);not null;default:'Yes'" json:"direction"`
	Chance       int16     `gorm:"type:smallint;not null" json:"chance"`
	NewAskPrice  string    `gorm:"type:numeric;not null;default:'0'" json:"new_ask_price"` // UINT256 mapped to string
	NewBidPrice  string    `gorm:"type:numeric;not null;default:'0'" json:"new_bid_price"` // UINT256 mapped to string
	Info         JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"info"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SubEventDirection) TableName() string {
	return "sub_event_direction"
}

type SubEventChanceStat struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	SubEventGUID string    `gorm:"type:varchar(500);not null;index:idx_sub_event_chance_stat_sub_event_guid" json:"sub_event_guid"`
	Chance       int16     `gorm:"type:smallint;not null" json:"chance"`
	Datetime     string    `gorm:"type:varchar(500);not null" json:"datetime"`
	StatWay      int16     `gorm:"type:smallint;not null" json:"stat_way"` // 0:1h, 1:6h, 2:1d, 3:1w, 4:All
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SubEventChanceStat) TableName() string {
	return "sub_event_chance_stat"
}

type EventView interface {
	GetEventByGUID(guid string) (*Event, error)
	GetEventByExternalID(externalID string) (*Event, error)
	QueryEvents(filter *Event) ([]Event, error)
	QueryEventsByCategory(categoryGUID string) ([]Event, error)
	QueryEventsByEcosystem(ecosystemGUID string) ([]Event, error)
	QueryEventsByDateRange(startDate, endDate time.Time) ([]Event, error)
	QueryEventsWithPagination(languageGUID, categoryGUID string, isLive *int16, page, limit int) ([]Event, int64, error)
	GetEventLanguage(eventGUID, languageGUID string) (*EventLanguage, error)
	QueryEventLanguages(eventGUID string) ([]EventLanguage, error)
	GetEventWithLanguage(eventGUID, languageGUID string) (*Event, *EventLanguage, error)
}

type EventDB interface {
	EventView
	CreateEvent(event *Event) error
	UpdateEvent(event *Event) error
	UpdateEventFields(eventGUID string, updates map[string]interface{}) error
	DeleteEvent(guid string) error
	CreateEventLanguage(eventLang *EventLanguage) error
	UpdateEventLanguage(eventLang *EventLanguage) error
	DeleteEventLanguage(guid string) error
}

type eventDB struct {
	gorm *gorm.DB
}

func NewEventDB(db *gorm.DB) EventDB {
	return &eventDB{gorm: db}
}

func (db *eventDB) GetEventByGUID(guid string) (*Event, error) {
	var event Event
	result := db.gorm.Where("guid = ?", guid).First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &event, nil
}

func (db *eventDB) GetEventByExternalID(externalID string) (*Event, error) {
	var event Event
	err := db.gorm.Where("external_id = ?", externalID).First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func (db *eventDB) QueryEvents(filter *Event) ([]Event, error) {
	var events []Event
	query := db.gorm.Model(&Event{})
	if filter != nil {
		if filter.CategoryGUID != "" {
			query = query.Where("category_guid = ?", filter.CategoryGUID)
		}
		if filter.EcosystemGUID != "" {
			query = query.Where("ecosystem_guid = ?", filter.EcosystemGUID)
		}
		if filter.EventPeriodGUID != "" {
			query = query.Where("event_period_guid = ?", filter.EventPeriodGUID)
		}
		if filter.IsLive >= 0 {
			query = query.Where("is_live = ?", filter.IsLive)
		}
		if filter.IsOnline {
			query = query.Where("is_online = ?", filter.IsOnline)
		}
		if filter.IsSports {
			query = query.Where("is_sports = ?", filter.IsSports)
		}
	}
	result := query.Order("created_at DESC").Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (db *eventDB) QueryEventsByCategory(categoryGUID string) ([]Event, error) {
	var events []Event
	result := db.gorm.Where("category_guid = ?", categoryGUID).
		Order("created_at DESC").
		Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (db *eventDB) QueryEventsByEcosystem(ecosystemGUID string) ([]Event, error) {
	var events []Event
	result := db.gorm.Where("ecosystem_guid = ?", ecosystemGUID).
		Order("created_at DESC").
		Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (db *eventDB) QueryEventsByDateRange(startDate, endDate time.Time) ([]Event, error) {
	var events []Event
	err := db.gorm.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (db *eventDB) QueryEventsWithPagination(languageGUID, categoryGUID string, isLive *int16, page, limit int) ([]Event, int64, error) {
	var events []Event
	var total int64

	// Build query
	query := db.gorm.Model(&Event{})

	// Filter by category
	if categoryGUID != "" {
		query = query.Where("category_guid = ?", categoryGUID)
	}

	// Filter by status
	if isLive != nil {
		query = query.Where("is_live = ?", *isLive)
	}

	// If language is specified, join event_language table
	if languageGUID != "" {
		query = query.Joins("INNER JOIN event_language ON event.guid = event_language.event_guid").
			Where("event_language.language_guid = ?", languageGUID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination query
	offset := (page - 1) * limit
	if err := query.Order("event.created_at DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (db *eventDB) CreateEvent(event *Event) error {
	return db.gorm.Create(event).Error
}

func (db *eventDB) UpdateEvent(event *Event) error {
	return db.gorm.Model(&Event{}).Where("guid = ?", event.GUID).Updates(event).Error
}

func (db *eventDB) UpdateEventFields(eventGUID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return db.gorm.Model(&Event{}).Where("guid = ?", eventGUID).Updates(updates).Error
}

func (db *eventDB) DeleteEvent(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&Event{}).Error
}

func (db *eventDB) GetEventLanguage(eventGUID, languageGUID string) (*EventLanguage, error) {
	var eventLang EventLanguage
	result := db.gorm.Where("event_guid = ? AND language_guid = ?", eventGUID, languageGUID).
		First(&eventLang)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &eventLang, nil
}

func (db *eventDB) QueryEventLanguages(eventGUID string) ([]EventLanguage, error) {
	var eventLangs []EventLanguage
	result := db.gorm.Where("event_guid = ?", eventGUID).
		Order("created_at ASC").
		Find(&eventLangs)
	if result.Error != nil {
		return nil, result.Error
	}
	return eventLangs, nil
}

func (db *eventDB) GetEventWithLanguage(eventGUID, languageGUID string) (*Event, *EventLanguage, error) {
	event, err := db.GetEventByGUID(eventGUID)
	if err != nil {
		return nil, nil, err
	}
	if event == nil {
		return nil, nil, nil
	}

	eventLang, err := db.GetEventLanguage(eventGUID, languageGUID)
	if err != nil {
		return event, nil, err
	}

	return event, eventLang, nil
}

func (db *eventDB) CreateEventLanguage(eventLang *EventLanguage) error {
	return db.gorm.Create(eventLang).Error
}

func (db *eventDB) UpdateEventLanguage(eventLang *EventLanguage) error {
	return db.gorm.Model(&EventLanguage{}).Where("guid = ?", eventLang.GUID).Updates(eventLang).Error
}

func (db *eventDB) DeleteEventLanguage(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&EventLanguage{}).Error
}
