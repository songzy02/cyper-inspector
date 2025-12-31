package model

import "time"

// UserRole 用户角色类型
type UserRole string

const (
	RoleAdmin UserRole = "admin" // 管理员
	RoleUser  UserRole = "user"  // 普通用户
)

// User 用户模型
type User struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;not null" json:"username"` // 用户名
	Password  string    `gorm:"size:255;not null" json:"-"`                   // 密码哈希
	Role      UserRole  `gorm:"size:20;default:user" json:"role"`             // 角色
	Enabled   bool      `gorm:"default:true" json:"enabled"`                  // 是否启用
	LastLogin time.Time `json:"last_login"`                                   // 最后登录时间
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// AgentToken Agent认证令牌
type AgentToken struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	AgentID   uint64    `gorm:"uniqueIndex;not null" json:"agent_id"` // Agent ID
	Token     string    `gorm:"size:255;not null" json:"token"`       // API Token
	Enabled   bool      `gorm:"default:true" json:"enabled"`          // 是否启用
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Agent     Agent     `gorm:"foreignKey:AgentID" json:"-"`
}

// TableName 表名
func (AgentToken) TableName() string {
	return "agent_tokens"
}

// LoginLog 登录日志
type LoginLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64;not null" json:"username"` // 用户名
	IP        string    `gorm:"size:15;not null" json:"ip"`       // 登录IP
	Status    string    `gorm:"size:20;not null" json:"status"`   // 登录状态
	Message   string    `gorm:"size:255" json:"message"`          // 登录消息
	UserAgent string    `gorm:"type:text" json:"user_agent"`      // User Agent
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (LoginLog) TableName() string {
	return "login_logs"
}
