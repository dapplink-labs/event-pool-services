package backend

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONB 自定义类型用于处理 PostgreSQL JSONB 字段
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
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

// Event 事件表
type Event struct {
	GUID                 string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	CategoryGUID         string    `gorm:"type:varchar(500);not null" json:"category_guid"`
	EcosystemGUID        string    `gorm:"type:varchar(500);not null" json:"ecosystem_guid"`
	EventPeriodGUID      string    `gorm:"type:varchar(500);not null" json:"event_period_guid"`
	MainTeamGroupGUID    string    `gorm:"type:varchar(255);not null" json:"main_team_group_guid"`
	ClusterTeamGroupGUID string    `gorm:"type:varchar(255);not null" json:"cluster_team_group_guid"`
	MainScore            string    `gorm:"type:numeric;not null" json:"main_score"`    // UINT256 mapped to string
	ClusterScore         string    `gorm:"type:numeric;not null" json:"cluster_score"` // UINT256 mapped to string
	Logo                 string    `gorm:"type:varchar(300);not null" json:"logo"`
	OrderType            int16     `gorm:"type:smallint;not null;default:0" json:"order_type"`
	OrderNum             string    `gorm:"type:numeric;not null" json:"order_num"` // UINT256 mapped to string
	OpenTime             string    `gorm:"type:varchar(100);not null" json:"open_time"`
	TradeVolume          float64   `gorm:"type:numeric(32,16);not null;default:0" json:"trade_volume"`
	ExperimentResult     string    `gorm:"type:text;not null" json:"experiment_result"`
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

// EventLanguage 事件多语言表
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

// SubEvent 事件子表
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

// SubEventLanguage 子事件多语言表
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

// SubEventDirection 子事件方向表
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

// SubEventChanceStat 子事件概率统计表
type SubEventChanceStat struct {
	GUID         string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	SubEventGUID string    `gorm:"type:varchar(500);not null;index:idx_sub_event_chance_stat_sub_event_guid" json:"sub_event_guid"`
	Chance       int16     `gorm:"type:smallint;not null" json:"chance"`
	Datetime     string    `gorm:"type:varchar(500);not null" json:"datetime"`
	StatWay      int16     `gorm:"type:smallint;not null" json:"stat_way"` // 0:1h; 2:6h; 3:1d; 4:1w; 5:All
	CreatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SubEventChanceStat) TableName() string {
	return "sub_event_chance_stat"
}
