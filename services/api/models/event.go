package models

// OutcomeRequest 结果选项请求
type OutcomeRequest struct {
	Name  string `json:"name" binding:"required"`  // 结果名称
	Color string `json:"color" binding:"required"` // 结果颜色
	Idx   int    `json:"idx" binding:"required"`   // 排序索引
}

// SubEventRequest 子事件请求
type SubEventRequest struct {
	Question string           `json:"question" binding:"required"` // 问题内容
	Outcomes []OutcomeRequest `json:"outcomes" binding:"required"` // 结果选项列表
}

// CreateEventRequest 创建事件请求
type CreateEventRequest struct {
	Title       string            `json:"title" binding:"required"`      // 事件标题
	Description string            `json:"description"`                   // 事件描述
	ImageURL    string            `json:"image_url"`                     // 事件图片URL
	StartDate   int64             `json:"start_date" binding:"required"` // 开始时间
	EndDate     int64             `json:"end_date" binding:"required"`   // 结束时间
	Tags        []string          `json:"tags"`                          // 标签列表
	SubEvents   []SubEventRequest `json:"sub_events" binding:"required"` // 子事件列表
}

// OutcomeResponse 结果选项响应
type OutcomeResponse struct {
	GUID         string `json:"guid"`           // 结果唯一标识
	SubEventGUID string `json:"sub_event_guid"` // 关联子事件GUID
	Name         string `json:"name"`           // 结果名称
	Color        string `json:"color"`          // 结果颜色
	Idx          int    `json:"idx"`            // 排序索引
	Created      int64  `json:"created"`        // 创建时间
	Updated      int64  `json:"updated"`        // 更新时间
}

// SubEventResponse 子事件响应
type SubEventResponse struct {
	GUID      string            `json:"guid"`       // 子事件唯一标识
	EventGUID string            `json:"event_guid"` // 关联事件GUID
	Question  string            `json:"question"`   // 问题内容
	Outcomes  []OutcomeResponse `json:"outcomes"`   // 结果选项列表
	Created   int64             `json:"created"`    // 创建时间
	Updated   int64             `json:"updated"`    // 更新时间
}

// TagResponse 标签响应
type TagResponse struct {
	GUID    string `json:"guid"`    // 标签唯一标识
	Name    string `json:"name"`    // 标签名称
	Created int64  `json:"created"` // 创建时间
	Updated int64  `json:"updated"` // 更新时间
}

// CreateEventResponse 创建事件响应
type CreateEventResponse struct {
	GUID        string             `json:"guid"`        // 事件唯一标识
	Title       string             `json:"title"`       // 事件标题
	Description string             `json:"description"` // 事件描述
	ImageURL    string             `json:"image_url"`   // 事件图片URL
	StartDate   int64              `json:"start_date"`  // 开始时间
	EndDate     int64              `json:"end_date"`    // 结束时间
	Tags        []TagResponse      `json:"tags"`        // 标签列表
	SubEvents   []SubEventResponse `json:"sub_events"`  // 子事件列表
	Created     int64              `json:"created"`     // 创建时间
	Updated     int64              `json:"updated"`     // 更新时间
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`             // 错误信息
	Message string `json:"message,omitempty"` // 详细信息
}

// ListEventsRequest 事件列表查询请求
type ListEventsRequest struct {
	Page      int      `json:"page"`       // 页码，默认1
	Limit     int      `json:"limit"`      // 每页数量，默认20，最大100
	Tags      []string `json:"tags"`       // 标签过滤（多个标签）
	Status    string   `json:"status"`     // 状态过滤：upcoming/active/ended
	Keyword   string   `json:"keyword"`    // 关键词搜索（标题/描述）
	StartTime int64    `json:"start_time"` // 时间范围起始
	EndTime   int64    `json:"end_time"`   // 时间范围结束
}

// EventListItem 事件列表项（基础信息）
type EventListItem struct {
	GUID        string   `json:"guid"`        // 事件唯一标识
	Title       string   `json:"title"`       // 事件标题
	Description string   `json:"description"` // 事件描述
	ImageURL    string   `json:"image_url"`   // 事件图片URL
	StartDate   int64    `json:"start_date"`  // 开始时间
	EndDate     int64    `json:"end_date"`    // 结束时间
	Status      string   `json:"status"`      // 状态：upcoming/active/ended
	Tags        []string `json:"tags"`        // 标签列表（仅名称）
	Created     int64    `json:"created"`     // 创建时间
	Updated     int64    `json:"updated"`     // 更新时间
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

// GetEventDetailResponse 事件详情响应（完整嵌套数据）
type GetEventDetailResponse struct {
	GUID        string             `json:"guid"`        // 事件唯一标识
	Title       string             `json:"title"`       // 事件标题
	Description string             `json:"description"` // 事件描述
	ImageURL    string             `json:"image_url"`   // 事件图片URL
	StartDate   int64              `json:"start_date"`  // 开始时间
	EndDate     int64              `json:"end_date"`    // 结束时间
	Status      string             `json:"status"`      // 状态：upcoming/active/ended
	Tags        []string           `json:"tags"`        // 标签列表（仅名称）
	SubEvents   []SubEventResponse `json:"sub_events"`  // 子事件列表
	Created     int64              `json:"created"`     // 创建时间
	Updated     int64              `json:"updated"`     // 更新时间
}
