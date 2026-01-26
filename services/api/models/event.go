package models

// ============================================
// 接口 A: 生成预测事件 (Create Event)
// ============================================
// SubEventDirectionRequest 子事件方向请求
type SubEventDirectionRequest struct {
	Direction string `json:"direction" binding:"required"` // 方向名称，例如 "Yes", "No"
	Chance    int16  `json:"chance" binding:"required"`    // 概率，例如 50
}

// SubEventRequest 子事件请求
type SubEventRequest struct {
	Title      string                     `json:"title" binding:"required"`      // 子事件标题
	Directions []SubEventDirectionRequest `json:"directions" binding:"required"` // 方向列表
}

// CreateEventRequest 创建事件请求
type CreateEventRequest struct {
	CategoryGUID         string            `json:"category_guid" binding:"required"`     // 分类 GUID
	EcosystemGUID        string            `json:"ecosystem_guid" binding:"required"`    // 生态 GUID
	EventPeriodGUID      string            `json:"event_period_guid" binding:"required"` // 时间标签 GUID
	MainTeamGroupGUID    string            `json:"main_team_group_guid"`                 // 主队 GUID（非运动类为空或"0"）
	ClusterTeamGroupGUID string            `json:"cluster_team_group_guid"`              // 客队 GUID（非运动类为空或"0"）
	Logo                 string            `json:"logo"`                                 // Logo URL
	Title                string            `json:"title" binding:"required"`             // 事件标题
	Rules                string            `json:"rules"`                                // 规则说明
	LanguageGUID         string            `json:"language_guid" binding:"required"`     // 语言 GUID
	SubEvents            []SubEventRequest `json:"sub_events" binding:"required,min=1"`  // 子事件列表
	IsSports             bool              `json:"is_sports"`                            // 是否为运动类事件
}

// SubEventDirectionResponse 子事件方向响应
type SubEventDirectionResponse struct {
	GUID        string `json:"guid"`          // 方向 GUID
	Direction   string `json:"direction"`     // 方向名称
	Chance      int16  `json:"chance"`        // 概率
	NewAskPrice string `json:"new_ask_price"` // 卖价
	NewBidPrice string `json:"new_bid_price"` // 买价
}

// SubEventResponse 子事件响应
type SubEventResponse struct {
	GUID       string                      `json:"guid"`       // 子事件 GUID
	Title      string                      `json:"title"`      // 子事件标题
	Logo       string                      `json:"logo"`       // Logo URL
	Directions []SubEventDirectionResponse `json:"directions"` // 方向列表
}

// CreateEventResponse 创建事件响应
type CreateEventResponse struct {
	GUID      string             `json:"guid"`       // 事件 GUID
	Title     string             `json:"title"`      // 事件标题
	Rules     string             `json:"rules"`      // 规则说明
	Logo      string             `json:"logo"`       // Logo URL
	SubEvents []SubEventResponse `json:"sub_events"` // 子事件列表
	CreatedAt string             `json:"created_at"` // 创建时间
}

// ============================================
// 接口 B: 查询事件列表 (List Events)
// ============================================

// ListEventsRequest 事件列表查询请求
type ListEventsRequest struct {
	Page         int    `json:"page"`          // 页码，默认1
	Limit        int    `json:"limit"`         // 每页数量，默认20，最大100
	LanguageGUID string `json:"language_guid"` // 语言 GUID（必需，用于多语言查询）
	CategoryGUID string `json:"category_guid"` // 分类 GUID（可选）
	IsLive       *int16 `json:"is_live"`       // 状态过滤：0-进行中, 1-预热, 2-已结束（可选）
}

// EventListItem 事件列表项
type EventListItem struct {
	GUID            string             `json:"guid"`              // 事件 GUID
	Title           string             `json:"title"`             // 事件标题（多语言）
	Rules           string             `json:"rules"`             // 规则说明（多语言）
	Logo            string             `json:"logo"`              // Logo URL
	CategoryGUID    string             `json:"category_guid"`     // 分类 GUID
	EcosystemGUID   string             `json:"ecosystem_guid"`    // 生态 GUID
	EventPeriodGUID string             `json:"event_period_guid"` // 时间标签 GUID
	IsLive          int16              `json:"is_live"`           // 状态：0-进行中, 1-预热, 2-已结束
	IsSports        bool               `json:"is_sports"`         // 是否为运动类事件
	OpenTime        string             `json:"open_time"`         // 开盘时间
	TradeVolume     float64            `json:"trade_volume"`      // 交易量
	SubEvents       []SubEventResponse `json:"sub_events"`        // 子事件列表（包含方向）
	CreatedAt       string             `json:"created_at"`        // 创建时间
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int `json:"page"`        // 当前页码
	Limit      int `json:"limit"`       // 每页数量
	Total      int `json:"total"`       // 总记录数
	TotalPages int `json:"total_pages"` // 总页数
}

// ListEventsResponse 事件列表响应
type ListEventsResponse struct {
	Events     []EventListItem `json:"events"`     // 事件列表
	Pagination PaginationInfo  `json:"pagination"` // 分页信息
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`             // 错误代码
	Message string `json:"message,omitempty"` // 错误详细信息
}
