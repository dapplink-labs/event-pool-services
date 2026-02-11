package database

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	Code      string    `gorm:"type:varchar(64)" json:"code"`
	SortOrder int       `gorm:"type:integer;not null;default:0" json:"sort_order"`
	IsActive  bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	Remark    string    `gorm:"type:varchar(200)" json:"remark"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Category) TableName() string {
	return "category"
}

type CategoryLanguage struct {
	GUID               string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LanguageGUID       string    `gorm:"type:varchar(255);not null" json:"language_guid"`
	CategoryGUID       string    `gorm:"type:varchar(255);not null" json:"category_guid"`
	ParentCategoryGUID string    `gorm:"type:varchar(255);not null" json:"parent_category_guid"`
	Level              int16     `gorm:"type:smallint;not null;default:0" json:"level"`
	Name               string    `gorm:"type:varchar(50);not null" json:"name"`
	Description        string    `gorm:"type:varchar(200);not null" json:"description"`
	CreatedAt          time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CategoryLanguage) TableName() string {
	return "category_language"
}

type CategoryView interface {
	GetCategoryByGUID(guid string) (*Category, error)
	GetCategoryByCode(code string) (*Category, error)
	QueryCategories(filter *Category) ([]Category, error)
	QueryActiveCategories() ([]Category, error)
	GetCategoryLanguage(categoryGUID, languageGUID string) (*CategoryLanguage, error)
	QueryCategoryLanguages(categoryGUID string) ([]CategoryLanguage, error)
}

type CategoryDB interface {
	CategoryView
	CreateCategory(category *Category) error
	UpdateCategory(category *Category) error
	DeleteCategory(guid string) error
	CreateCategoryLanguage(categoryLang *CategoryLanguage) error
	UpdateCategoryLanguage(categoryLang *CategoryLanguage) error
	DeleteCategoryLanguage(guid string) error
}

type categoryDB struct {
	gorm *gorm.DB
}

func NewCategoryDB(db *gorm.DB) CategoryDB {
	return &categoryDB{gorm: db}
}

func (db *categoryDB) GetCategoryByGUID(guid string) (*Category, error) {
	var category Category
	result := db.gorm.Where("guid = ?", guid).First(&category)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &category, nil
}

func (db *categoryDB) GetCategoryByCode(code string) (*Category, error) {
	var category Category
	result := db.gorm.Where("code = ?", code).First(&category)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &category, nil
}

func (db *categoryDB) QueryCategories(filter *Category) ([]Category, error) {
	var categories []Category
	query := db.gorm.Model(&Category{})
	if filter != nil {
		if filter.Code != "" {
			query = query.Where("code = ?", filter.Code)
		}
		if filter.IsActive {
			query = query.Where("is_active = ?", filter.IsActive)
		}
	}
	result := query.Order("sort_order ASC, created_at DESC").Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

func (db *categoryDB) QueryActiveCategories() ([]Category, error) {
	var categories []Category
	result := db.gorm.Where("is_active = ?", true).
		Order("sort_order ASC, created_at DESC").
		Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

func (db *categoryDB) CreateCategory(category *Category) error {
	return db.gorm.Create(category).Error
}

func (db *categoryDB) UpdateCategory(category *Category) error {
	return db.gorm.Model(&Category{}).Where("guid = ?", category.GUID).Updates(category).Error
}

func (db *categoryDB) DeleteCategory(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&Category{}).Error
}

func (db *categoryDB) GetCategoryLanguage(categoryGUID, languageGUID string) (*CategoryLanguage, error) {
	var categoryLang CategoryLanguage
	result := db.gorm.Where("category_guid = ? AND language_guid = ?", categoryGUID, languageGUID).
		First(&categoryLang)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &categoryLang, nil
}

func (db *categoryDB) QueryCategoryLanguages(categoryGUID string) ([]CategoryLanguage, error) {
	var categoryLangs []CategoryLanguage
	result := db.gorm.Where("category_guid = ?", categoryGUID).
		Order("created_at ASC").
		Find(&categoryLangs)
	if result.Error != nil {
		return nil, result.Error
	}
	return categoryLangs, nil
}

func (db *categoryDB) CreateCategoryLanguage(categoryLang *CategoryLanguage) error {
	return db.gorm.Create(categoryLang).Error
}

func (db *categoryDB) UpdateCategoryLanguage(categoryLang *CategoryLanguage) error {
	return db.gorm.Model(&CategoryLanguage{}).Where("guid = ?", categoryLang.GUID).Updates(categoryLang).Error
}

func (db *categoryDB) DeleteCategoryLanguage(guid string) error {
	return db.gorm.Where("guid = ?", guid).Delete(&CategoryLanguage{}).Error
}
