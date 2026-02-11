package database

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Ecosystem struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	CategoryGUID string    `gorm:"type:varchar(255);not null" json:"category_guid"`
	EventNum     string    `gorm:"type:numeric;not null" json:"event_num"` // UINT256 mapped to string for large numbers
	Code         string    `gorm:"type:varchar(64)" json:"code"`
	SortOrder    int       `gorm:"type:integer;not null;default:0" json:"sort_order"`
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	Remark       string    `gorm:"type:varchar(200)" json:"remark"`
	Extra        JSONB     `gorm:"type:jsonb" json:"extra"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Ecosystem) TableName() string {
	return "ecosystem"
}

type EcosystemLanguage struct {
	GUID          string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LanguageGUID  string    `gorm:"type:varchar(255);not null" json:"language_guid"`
	EcosystemGUID string    `gorm:"type:varchar(255);not null" json:"ecosystem_guid"`
	Name          string    `gorm:"type:varchar(50);not null" json:"name"`
	Description   string    `gorm:"type:varchar(200);not null" json:"description"`
	CreatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EcosystemLanguage) TableName() string {
	return "ecosystem_language"
}

type EcosystemView interface {
	GetEcosystemByGUID(guid string) (*Ecosystem, error)
	GetEcosystemByCode(code string) (*Ecosystem, error)
	QueryEcosystems(filter *Ecosystem) ([]Ecosystem, error)
	QueryEcosystemsByCategory(categoryGUID string) ([]Ecosystem, error)
	QueryActiveEcosystems() ([]Ecosystem, error)
	GetEcosystemLanguage(ecosystemGUID, languageGUID string) (*EcosystemLanguage, error)
	QueryEcosystemLanguages(ecosystemGUID string) ([]EcosystemLanguage, error)
}

type EcosystemDB interface {
	EcosystemView
	CreateEcosystem(ecosystem *Ecosystem) error
	UpdateEcosystem(ecosystem *Ecosystem) error
	DeleteEcosystem(guid string) error
	CreateEcosystemLanguage(ecosystemLang *EcosystemLanguage) error
	UpdateEcosystemLanguage(ecosystemLang *EcosystemLanguage) error
	DeleteEcosystemLanguage(guid string) error
}

type ecosystemDB struct {
	gorm *gorm.DB
}

func NewEcosystemDB(db *gorm.DB) EcosystemDB {
	return &ecosystemDB{gorm: db}
}

func (db *ecosystemDB) GetEcosystemByGUID(guid string) (*Ecosystem, error) {
	var ecosystem Ecosystem
	result := db.gorm.Where("guid = ?", guid).First(&ecosystem)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &ecosystem, nil
}

func (db *ecosystemDB) GetEcosystemByCode(code string) (*Ecosystem, error) {
	var ecosystem Ecosystem
	result := db.gorm.Where("code = ?", code).First(&ecosystem)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &ecosystem, nil
}

func (db *ecosystemDB) QueryEcosystems(filter *Ecosystem) ([]Ecosystem, error) {
	var ecosystems []Ecosystem
	query := db.gorm.Model(&Ecosystem{})
	if filter != nil {
		if filter.Code != "" {
			query = query.Where("code = ?", filter.Code)
		}
		if filter.CategoryGUID != "" {
			query = query.Where("category_guid = ?", filter.CategoryGUID)
		}
		if filter.IsActive {
			query = query.Where("is_active = ?", filter.IsActive)
		}
	}
	result := query.Order("sort_order ASC, created_at DESC").Find(&ecosystems)
	if result.Error != nil {
		return nil, result.Error
	}
	return ecosystems, nil
}

func (db *ecosystemDB) QueryEcosystemsByCategory(categoryGUID string) ([]Ecosystem, error) {
	var ecosystems []Ecosystem
	result := db.gorm.Where("category_guid = ? AND is_active = ?", categoryGUID, true).
		Order("sort_order ASC, created_at DESC").
		Find(&ecosystems)
	if result.Error != nil {
		return nil, result.Error
	}
	return ecosystems, nil
}

func (db *ecosystemDB) QueryActiveEcosystems() ([]Ecosystem, error) {
	var ecosystems []Ecosystem
	result := db.gorm.Where("is_active = ?", true).
		Order("sort_order ASC, created_at DESC").
		Find(&ecosystems)
	if result.Error != nil {
		return nil, result.Error
	}
	return ecosystems, nil
}

func (db *ecosystemDB) CreateEcosystem(ecosystem *Ecosystem) error {
	return db.gorm.Create(ecosystem).Error
}

func (db *ecosystemDB) UpdateEcosystem(ecosystem *Ecosystem) error {
	return db.gorm.Model(&Ecosystem{}).Where("guid = ?", ecosystem.GUID).Updates(ecosystem).Error
}

func (db *ecosystemDB) DeleteEcosystem(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&Ecosystem{}).Error
}

func (db *ecosystemDB) GetEcosystemLanguage(ecosystemGUID, languageGUID string) (*EcosystemLanguage, error) {
	var ecosystemLang EcosystemLanguage
	result := db.gorm.Where("ecosystem_guid = ? AND language_guid = ?", ecosystemGUID, languageGUID).
		First(&ecosystemLang)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &ecosystemLang, nil
}

func (db *ecosystemDB) QueryEcosystemLanguages(ecosystemGUID string) ([]EcosystemLanguage, error) {
	var ecosystemLangs []EcosystemLanguage
	result := db.gorm.Where("ecosystem_guid = ?", ecosystemGUID).
		Order("created_at ASC").
		Find(&ecosystemLangs)
	if result.Error != nil {
		return nil, result.Error
	}
	return ecosystemLangs, nil
}

func (db *ecosystemDB) CreateEcosystemLanguage(ecosystemLang *EcosystemLanguage) error {
	return db.gorm.Create(ecosystemLang).Error
}

func (db *ecosystemDB) UpdateEcosystemLanguage(ecosystemLang *EcosystemLanguage) error {
	return db.gorm.Model(&EcosystemLanguage{}).Where("guid = ?", ecosystemLang.GUID).Updates(ecosystemLang).Error
}

func (db *ecosystemDB) DeleteEcosystemLanguage(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&EcosystemLanguage{}).Error
}
