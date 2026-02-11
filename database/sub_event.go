package database

import (
	"errors"

	"gorm.io/gorm"
)

type SubEventView interface {
	GetSubEventByGUID(guid string) (*SubEvent, error)
	QuerySubEventsByEventGUID(eventGUID string) ([]SubEvent, error)
	GetSubEventLanguage(subEventGUID, languageGUID string) (*SubEventLanguage, error)
	QuerySubEventLanguages(subEventGUID string) ([]SubEventLanguage, error)
	GetSubEventDirectionByGUID(guid string) (*SubEventDirection, error)
	QuerySubEventDirections(subEventGUID string) ([]SubEventDirection, error)
	GetSubEventChanceStatByGUID(guid string) (*SubEventChanceStat, error)
	QuerySubEventChanceStats(subEventGUID string) ([]SubEventChanceStat, error)
}

type SubEventDB interface {
	SubEventView
	CreateSubEvent(subEvent *SubEvent) error
	UpdateSubEvent(subEvent *SubEvent) error
	DeleteSubEvent(guid string) error
	CreateSubEventLanguage(subEventLang *SubEventLanguage) error
	UpdateSubEventLanguage(subEventLang *SubEventLanguage) error
	DeleteSubEventLanguage(guid string) error
	CreateSubEventDirection(direction *SubEventDirection) error
	UpdateSubEventDirection(direction *SubEventDirection) error
	DeleteSubEventDirection(guid string) error
	CreateSubEventChanceStat(stat *SubEventChanceStat) error
	UpdateSubEventChanceStat(stat *SubEventChanceStat) error
	DeleteSubEventChanceStat(guid string) error
}

type subEventDB struct {
	gorm *gorm.DB
}

func NewSubEventDB(db *gorm.DB) SubEventDB {
	return &subEventDB{gorm: db}
}

func (db *subEventDB) GetSubEventByGUID(guid string) (*SubEvent, error) {
	var subEvent SubEvent
	result := db.gorm.Where("guid = ?", guid).First(&subEvent)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &subEvent, nil
}

func (db *subEventDB) QuerySubEventsByEventGUID(eventGUID string) ([]SubEvent, error) {
	var subEvents []SubEvent
	if err := db.gorm.Where("parent_event_guid = ?", eventGUID).
		Order("created_at ASC").
		Find(&subEvents).Error; err != nil {
		return nil, err
	}
	return subEvents, nil
}

func (db *subEventDB) CreateSubEvent(subEvent *SubEvent) error {
	return db.gorm.Create(subEvent).Error
}

func (db *subEventDB) UpdateSubEvent(subEvent *SubEvent) error {
	return db.gorm.Model(&SubEvent{}).Where("guid = ?", subEvent.GUID).Updates(subEvent).Error
}

func (db *subEventDB) DeleteSubEvent(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&SubEvent{}).Error
}

func (db *subEventDB) GetSubEventLanguage(subEventGUID, languageGUID string) (*SubEventLanguage, error) {
	var subEventLang SubEventLanguage
	result := db.gorm.Where("sub_event_guid = ? AND language_guid = ?", subEventGUID, languageGUID).
		First(&subEventLang)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &subEventLang, nil
}

func (db *subEventDB) QuerySubEventLanguages(subEventGUID string) ([]SubEventLanguage, error) {
	var subEventLangs []SubEventLanguage
	result := db.gorm.Where("sub_event_guid = ?", subEventGUID).
		Order("created_at ASC").
		Find(&subEventLangs)
	if result.Error != nil {
		return nil, result.Error
	}
	return subEventLangs, nil
}

func (db *subEventDB) CreateSubEventLanguage(subEventLang *SubEventLanguage) error {
	return db.gorm.Create(subEventLang).Error
}

func (db *subEventDB) UpdateSubEventLanguage(subEventLang *SubEventLanguage) error {
	return db.gorm.Model(&SubEventLanguage{}).Where("guid = ?", subEventLang.GUID).Updates(subEventLang).Error
}

func (db *subEventDB) DeleteSubEventLanguage(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&SubEventLanguage{}).Error
}

func (db *subEventDB) GetSubEventDirectionByGUID(guid string) (*SubEventDirection, error) {
	var direction SubEventDirection
	result := db.gorm.Where("guid = ?", guid).First(&direction)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &direction, nil
}

func (db *subEventDB) QuerySubEventDirections(subEventGUID string) ([]SubEventDirection, error) {
	var directions []SubEventDirection
	if err := db.gorm.Where("sub_event_guid = ?", subEventGUID).
		Order("created_at ASC").
		Find(&directions).Error; err != nil {
		return nil, err
	}
	return directions, nil
}

func (db *subEventDB) CreateSubEventDirection(direction *SubEventDirection) error {
	return db.gorm.Create(direction).Error
}

func (db *subEventDB) UpdateSubEventDirection(direction *SubEventDirection) error {
	return db.gorm.Model(&SubEventDirection{}).Where("guid = ?", direction.GUID).Updates(direction).Error
}

func (db *subEventDB) DeleteSubEventDirection(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&SubEventDirection{}).Error
}

func (db *subEventDB) GetSubEventChanceStatByGUID(guid string) (*SubEventChanceStat, error) {
	var stat SubEventChanceStat
	result := db.gorm.Where("guid = ?", guid).First(&stat)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &stat, nil
}

func (db *subEventDB) QuerySubEventChanceStats(subEventGUID string) ([]SubEventChanceStat, error) {
	var stats []SubEventChanceStat
	if err := db.gorm.Where("sub_event_guid = ?", subEventGUID).
		Order("created_at ASC").
		Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (db *subEventDB) CreateSubEventChanceStat(stat *SubEventChanceStat) error {
	return db.gorm.Create(stat).Error
}

func (db *subEventDB) UpdateSubEventChanceStat(stat *SubEventChanceStat) error {
	return db.gorm.Model(&SubEventChanceStat{}).Where("guid = ?", stat.GUID).Updates(stat).Error
}

func (db *subEventDB) DeleteSubEventChanceStat(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&SubEventChanceStat{}).Error
}
