package database

import "time"

// SysLog 系统日志表
type SysLog struct {
	GUID        string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	Action      string    `gorm:"type:varchar(100);default:''" json:"action"`
	Desc        string    `gorm:"column:desc;type:varchar(100);default:''" json:"desc"`
	Admin       string    `gorm:"type:varchar(30);default:''" json:"admin"`
	IP          string    `gorm:"type:varchar(30);default:''" json:"ip"`
	Cate        int16     `gorm:"type:smallint;default:0" json:"cate"`
	Status      int16     `gorm:"type:smallint;default:-1" json:"status"`
	Asset       string    `gorm:"type:varchar(255);default:''" json:"asset"`
	Before      string    `gorm:"type:varchar(255);default:''" json:"before"`
	After       string    `gorm:"type:varchar(255);default:''" json:"after"`
	UserID      int64     `gorm:"type:bigint;default:0" json:"user_id"`
	OrderNumber string    `gorm:"type:varchar(64);default:''" json:"order_number"`
	Op          int16     `gorm:"type:smallint;default:-1" json:"op"`
	CreatedAt   time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysLog) TableName() string {
	return "sys_log"
}

// Auth 权限表
type Auth struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	AuthName  string    `gorm:"type:varchar(255);default:''" json:"auth_name"`
	AuthURL   string    `gorm:"type:varchar(255);default:''" json:"auth_url"`
	UserID    int       `gorm:"type:int;default:0" json:"user_id"`
	PID       int       `gorm:"type:int;default:0" json:"pid"`
	Sort      int       `gorm:"type:int;default:0" json:"sort"`
	Icon      string    `gorm:"type:varchar(255);default:''" json:"icon"`
	IsShow    int       `gorm:"type:int;default:1" json:"is_show"`
	Status    int       `gorm:"type:int;default:1" json:"status"`
	CreateID  int       `gorm:"type:int;default:0" json:"create_id"`
	UpdateID  int       `gorm:"type:int;default:0" json:"update_id"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Auth) TableName() string {
	return "auth"
}

// Role 角色表
type Role struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	RoleName  string    `gorm:"type:varchar(100);default:''" json:"role_name"`
	Detail    string    `gorm:"type:varchar(255);default:''" json:"detail"`
	Status    int       `gorm:"type:int;default:1" json:"status"`
	CreateID  int       `gorm:"type:int;default:0" json:"create_id"`
	UpdateID  int       `gorm:"type:int;default:0" json:"update_id"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Role) TableName() string {
	return "role"
}

// RoleAuth 角色权限关联表
type RoleAuth struct {
	AuthID int   `gorm:"type:int;not null;primaryKey" json:"auth_id"`
	RoleID int64 `gorm:"type:bigint;not null;primaryKey" json:"role_id"`
}

func (RoleAuth) TableName() string {
	return "role_auth"
}

// Admin 管理员表
type Admin struct {
	GUID      string    `gorm:"type:text;primaryKey;default:replace(uuid_generate_v4()::text, '-', '')" json:"guid"`
	LoginName string    `gorm:"type:varchar(32);not null;uniqueIndex" json:"login_name"`
	RealName  string    `gorm:"type:varchar(32);uniqueIndex" json:"real_name"`
	Password  string    `gorm:"type:varchar(100);not null" json:"password"`
	RoleIDs   string    `gorm:"type:varchar(255);default:''" json:"role_ids"`
	Phone     string    `gorm:"type:varchar(11);uniqueIndex" json:"phone"`
	Email     string    `gorm:"type:varchar(32)" json:"email"`
	Salt      string    `gorm:"type:varchar(255);default:''" json:"salt"`
	LastLogin int64     `gorm:"type:bigint;default:0" json:"last_login"`
	LastIP    string    `gorm:"type:varchar(255);default:''" json:"last_ip"`
	Status    int       `gorm:"type:int;default:1" json:"status"`
	CreateID  int       `gorm:"type:int;default:0" json:"create_id"`
	UpdateID  int       `gorm:"type:int;default:0" json:"update_id"`
	CreatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Admin) TableName() string {
	return "admin"
}
