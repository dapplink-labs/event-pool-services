package database

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type EventPeriod struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	Code      string    `gorm:"type:varchar(64)" json:"code"`
	IsActive  bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	Scheduled string    `gorm:"type:varchar(64)" json:"scheduled"`
	Remark    string    `gorm:"type:varchar(200)" json:"remark"`
	Extra     JSONB     `gorm:"type:jsonb" json:"extra"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EventPeriod) TableName() string {
	return "event_period"
}

type EventPeriodLanguage struct {
	GUID            string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	EventPeriodGUID string    `gorm:"type:varchar(255);not null" json:"event_period_guid"`
	LanguageGUID    string    `gorm:"type:varchar(255);not null" json:"language_guid"`
	Name            string    `gorm:"type:varchar(50);not null" json:"name"`
	Description     string    `gorm:"type:varchar(200);not null" json:"description"`
	CreatedAt       time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EventPeriodLanguage) TableName() string {
	return "event_period_language"
}

type EventPeriodView interface {
	GetEventPeriodByGUID(guid string) (*EventPeriod, error)
	GetEventPeriodByCode(code string) (*EventPeriod, error)
	GetEventPeriodByEventGUID(eventGUID string) (*EventPeriod, error)
	QueryEventPeriods(filter *EventPeriod) ([]EventPeriod, error)
	QueryEventPeriodsByEvent(eventGUID string) ([]EventPeriod, error)
	QueryActiveEventPeriods() ([]EventPeriod, error)
	GetEventPeriodLanguage(eventPeriodGUID, languageGUID string) (*EventPeriodLanguage, error)
	QueryEventPeriodLanguages(eventPeriodGUID string) ([]EventPeriodLanguage, error)
}

type EventPeriodDB interface {
	EventPeriodView
	CreateEventPeriod(eventPeriod *EventPeriod) error
	UpdateEventPeriod(eventPeriod *EventPeriod) error
	DeleteEventPeriod(guid string) error
	CreateEventPeriodLanguage(eventPeriodLang *EventPeriodLanguage) error
	UpdateEventPeriodLanguage(eventPeriodLang *EventPeriodLanguage) error
	DeleteEventPeriodLanguage(guid string) error
}

type eventPeriodDB struct {
	gorm *gorm.DB
}

func NewEventPeriodDB(db *gorm.DB) EventPeriodDB {
	return &eventPeriodDB{gorm: db}
}

func (db *eventPeriodDB) GetEventPeriodByGUID(guid string) (*EventPeriod, error) {
	var eventPeriod EventPeriod
	result := db.gorm.Where("guid = ?", guid).First(&eventPeriod)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &eventPeriod, nil
}

func (db *eventPeriodDB) GetEventPeriodByCode(code string) (*EventPeriod, error) {
	var eventPeriod EventPeriod
	result := db.gorm.Where("code = ?", code).First(&eventPeriod)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &eventPeriod, nil
}

func (db *eventPeriodDB) GetEventPeriodByEventGUID(eventGUID string) (*EventPeriod, error) {
	var eventPeriod EventPeriod
	result := db.gorm.Where("event_guid = ?", eventGUID).First(&eventPeriod)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &eventPeriod, nil
}

func (db *eventPeriodDB) QueryEventPeriods(filter *EventPeriod) ([]EventPeriod, error) {
	var eventPeriods []EventPeriod
	query := db.gorm.Model(&EventPeriod{})
	if filter != nil {
		if filter.Code != "" {
			query = query.Where("code = ?", filter.Code)
		}
		if filter.IsActive {
			query = query.Where("is_active = ?", filter.IsActive)
		}
	}
	result := query.Order("created_at DESC").Find(&eventPeriods)
	if result.Error != nil {
		return nil, result.Error
	}
	return eventPeriods, nil
}

func (db *eventPeriodDB) QueryEventPeriodsByEvent(eventGUID string) ([]EventPeriod, error) {
	var eventPeriods []EventPeriod
	result := db.gorm.Where("event_guid = ? AND is_active = ?", eventGUID, true).
		Order("created_at DESC").
		Find(&eventPeriods)
	if result.Error != nil {
		return nil, result.Error
	}
	return eventPeriods, nil
}

func (db *eventPeriodDB) QueryActiveEventPeriods() ([]EventPeriod, error) {
	var eventPeriods []EventPeriod
	result := db.gorm.Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&eventPeriods)
	if result.Error != nil {
		return nil, result.Error
	}
	return eventPeriods, nil
}

func (db *eventPeriodDB) CreateEventPeriod(eventPeriod *EventPeriod) error {
	return db.gorm.Create(eventPeriod).Error
}

func (db *eventPeriodDB) UpdateEventPeriod(eventPeriod *EventPeriod) error {
	return db.gorm.Model(&EventPeriod{}).Where("guid = ?", eventPeriod.GUID).Updates(eventPeriod).Error
}

func (db *eventPeriodDB) DeleteEventPeriod(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&EventPeriod{}).Error
}

func (db *eventPeriodDB) GetEventPeriodLanguage(eventPeriodGUID, languageGUID string) (*EventPeriodLanguage, error) {
	var eventPeriodLang EventPeriodLanguage
	result := db.gorm.Where("event_period_guid = ? AND language_guid = ?", eventPeriodGUID, languageGUID).
		First(&eventPeriodLang)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &eventPeriodLang, nil
}

func (db *eventPeriodDB) QueryEventPeriodLanguages(eventPeriodGUID string) ([]EventPeriodLanguage, error) {
	var eventPeriodLangs []EventPeriodLanguage
	result := db.gorm.Where("event_period_guid = ?", eventPeriodGUID).
		Order("created_at ASC").
		Find(&eventPeriodLangs)
	if result.Error != nil {
		return nil, result.Error
	}
	return eventPeriodLangs, nil
}

func (db *eventPeriodDB) CreateEventPeriodLanguage(eventPeriodLang *EventPeriodLanguage) error {
	return db.gorm.Create(eventPeriodLang).Error
}

func (db *eventPeriodDB) UpdateEventPeriodLanguage(eventPeriodLang *EventPeriodLanguage) error {
	return db.gorm.Model(&EventPeriodLanguage{}).Where("guid = ?", eventPeriodLang.GUID).Updates(eventPeriodLang).Error
}

func (db *eventPeriodDB) DeleteEventPeriodLanguage(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&EventPeriodLanguage{}).Error
}
