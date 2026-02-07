package database

import "time"

// Languages 支持的语言表
type Languages struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LanguageName string    `gorm:"type:varchar;default:'zh'" json:"language_name"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Languages) TableName() string {
	return "languages"
}

// Category 事件分类表
type Category struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	IsActive  bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Category) TableName() string {
	return "category"
}

// CategoryLanguage 事件分类多语言表
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

// Ecosystem 所属生态
type Ecosystem struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	CategoryGUID string    `gorm:"type:varchar(255);not null" json:"category_guid"`
	EventNum     string    `gorm:"type:numeric;not null" json:"event_num"` // UINT256 mapped to string for large numbers
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Ecosystem) TableName() string {
	return "ecosystem"
}

// EcosystemLanguage 所属生态多语言表
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

// EventPeriod 时间标签表
type EventPeriod struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	CategoryGUID string    `gorm:"type:varchar(255);not null" json:"category_guid"`
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EventPeriod) TableName() string {
	return "event_period"
}

// EventPeriodLanguage 时间标签表多语言表
type EventPeriodLanguage struct {
	GUID            string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	EventPeriodGUID string    `gorm:"type:varchar(255);not null" json:"event_period_guid"`
	LanguageGUID    string    `gorm:"type:varchar(255);not null" json:"language_guid"`
	Name            string    `gorm:"type:varchar(50);not null" json:"name"`
	CreatedAt       time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (EventPeriodLanguage) TableName() string {
	return "event_period_language"
}

// TeamGroup 运动类团队
type TeamGroup struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	Logo      string    `gorm:"type:varchar(255);not null" json:"logo"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TeamGroup) TableName() string {
	return "team_group"
}

// TeamGroupLanguage 运动类团队多语言表
type TeamGroupLanguage struct {
	GUID          string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	TeamGroupGUID string    `gorm:"type:varchar(255);not null" json:"team_group_guid"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TeamGroupLanguage) TableName() string {
	return "team_group_language"
}
