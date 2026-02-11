package database

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type TeamGroup struct {
	GUID          string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	EcosystemGUID string    `gorm:"type:varchar(255);not null" json:"ecosystem_guid"`
	ExternalId    string    `gorm:"type:varchar(255);not null;default:'0'" json:"external_id"`
	Logo          string    `gorm:"type:varchar(255);not null" json:"logo"`
	CreatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TeamGroup) TableName() string {
	return "team_group"
}

type TeamGroupLanguage struct {
	GUID          string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LanguageGUID  string    `gorm:"type:varchar(255);not null" json:"language_guid"`
	TeamGroupGUID string    `gorm:"type:varchar(255);not null" json:"team_group_guid"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Alias         string    `gorm:"type:varchar(255);not null" json:"alias"`
	CreatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TeamGroupLanguage) TableName() string {
	return "team_group_language"
}

type TeamGroupView interface {
	GetTeamGroupByGUID(guid string) (*TeamGroup, error)
	QueryTeamGroups(filter *TeamGroup) ([]TeamGroup, error)
	GetTeamGroupLanguage(teamGroupGUID, languageGUID string) (*TeamGroupLanguage, error)
	QueryTeamGroupLanguages(teamGroupGUID string) ([]TeamGroupLanguage, error)
	GetTeamGroupByExternalId(externalId string, ecosystem string) (*TeamGroup, error)
}

type TeamGroupDB interface {
	TeamGroupView
	CreateTeamGroup(teamGroup *TeamGroup) error
	UpdateTeamGroup(teamGroup *TeamGroup) error
	DeleteTeamGroup(guid string) error
	CreateTeamGroupLanguage(teamGroupLang *TeamGroupLanguage) error
	UpdateTeamGroupLanguage(teamGroupLang *TeamGroupLanguage) error
	DeleteTeamGroupLanguage(guid string) error
}

type teamGroupDB struct {
	gorm *gorm.DB
}

func NewTeamGroupDB(db *gorm.DB) TeamGroupDB {
	return &teamGroupDB{gorm: db}
}

func (db *teamGroupDB) GetTeamGroupByGUID(guid string) (*TeamGroup, error) {
	var teamGroup TeamGroup
	result := db.gorm.Where("guid = ?", guid).First(&teamGroup)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &teamGroup, nil
}

func (db *teamGroupDB) QueryTeamGroups(filter *TeamGroup) ([]TeamGroup, error) {
	var teamGroups []TeamGroup
	query := db.gorm.Model(&TeamGroup{})
	if filter != nil {
		if filter.Logo != "" {
			query = query.Where("logo = ?", filter.Logo)
		}
	}
	result := query.Order("created_at DESC").Find(&teamGroups)
	if result.Error != nil {
		return nil, result.Error
	}
	return teamGroups, nil
}

func (db *teamGroupDB) GetTeamGroupByExternalId(externalId string, ecosystem string) (*TeamGroup, error) {
	var teamGroup TeamGroup
	result := db.gorm.Where("external_id = ? AND ecosystem_guid = ?", externalId, ecosystem).
		First(&teamGroup)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &teamGroup, nil
}

func (db *teamGroupDB) CreateTeamGroup(teamGroup *TeamGroup) error {
	return db.gorm.Create(teamGroup).Error
}

func (db *teamGroupDB) UpdateTeamGroup(teamGroup *TeamGroup) error {
	return db.gorm.Model(&TeamGroup{}).Where("guid = ?", teamGroup.GUID).Updates(teamGroup).Error
}

func (db *teamGroupDB) DeleteTeamGroup(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&TeamGroup{}).Error
}

func (db *teamGroupDB) GetTeamGroupLanguage(teamGroupGUID, languageGUID string) (*TeamGroupLanguage, error) {
	var teamGroupLang TeamGroupLanguage
	result := db.gorm.Where("team_group_guid = ? AND language_guid = ?", teamGroupGUID, languageGUID).
		First(&teamGroupLang)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &teamGroupLang, nil
}

func (db *teamGroupDB) QueryTeamGroupLanguages(teamGroupGUID string) ([]TeamGroupLanguage, error) {
	var teamGroupLangs []TeamGroupLanguage
	result := db.gorm.Where("team_group_guid = ?", teamGroupGUID).
		Order("created_at ASC").
		Find(&teamGroupLangs)
	if result.Error != nil {
		return nil, result.Error
	}
	return teamGroupLangs, nil
}

func (db *teamGroupDB) CreateTeamGroupLanguage(teamGroupLang *TeamGroupLanguage) error {
	return db.gorm.Create(teamGroupLang).Error
}

func (db *teamGroupDB) UpdateTeamGroupLanguage(teamGroupLang *TeamGroupLanguage) error {
	return db.gorm.Model(&TeamGroupLanguage{}).Where("guid = ?", teamGroupLang.GUID).Updates(teamGroupLang).Error
}

func (db *teamGroupDB) DeleteTeamGroupLanguage(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&TeamGroupLanguage{}).Error
}
