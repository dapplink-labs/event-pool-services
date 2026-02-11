package database

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Languages struct {
	GUID          string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LanguageName  string    `gorm:"type:varchar(20);not null;unique" json:"language_name"`
	LanguageLabel string    `gorm:"type:varchar(50)" json:"language_label"`
	IsDefault     bool      `gorm:"type:boolean;not null;default:false" json:"is_default"`
	IsActive      bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Languages) TableName() string {
	return "languages"
}

type LanguagesView interface {
	QueryAllLanguages() (lR []Languages, err error)
}

type LanguagesDB interface {
	LanguagesView
}

type languagesDB struct {
	gorm *gorm.DB
}

func NewLanguagesDB(db *gorm.DB) LanguagesDB {
	return &languagesDB{gorm: db}
}

func (db *languagesDB) QueryAllLanguages() (lR []Languages, err error) {
	var languagesRecords []Languages
	result := db.gorm.Table("languages").Find(&languagesRecords)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return languagesRecords, nil
}
