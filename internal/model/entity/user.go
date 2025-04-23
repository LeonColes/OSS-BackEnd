package entity

import (
	"time"
)

// User 用户实体
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`                       // 用户ID
	Email        string     `gorm:"size:100;not null;uniqueIndex" json:"email"` // 用户邮箱，登录凭证
	Name         string     `gorm:"size:50;not null" json:"name"`               // 用户姓名
	PasswordHash string     `gorm:"size:100;not null" json:"-"`                 // 密码哈希值，不返回给前端
	Avatar       string     `gorm:"size:255" json:"avatar"`                     // 用户头像URL
	Status       int        `gorm:"default:1" json:"status"`                    // 状态（1-正常，2-禁用，3-锁定）
	LastLoginAt  *time.Time `json:"last_login_at"`                              // 最后登录时间
	LastLoginIP  string     `gorm:"size:50" json:"last_login_ip"`               // 最后登录IP
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`           // 创建时间
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`           // 更新时间

	// 用户角色关联（多对多）
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"` // 用户角色
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// 用户状态常量
const (
	UserStatusNormal   = 1 // 正常状态
	UserStatusDisabled = 2 // 禁用状态
	UserStatusLocked   = 3 // 锁定状态
)

// 隐藏敏感信息
func (u *User) HideSensitiveInfo() {
	u.PasswordHash = ""
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID    uint      `gorm:"primaryKey;column:user_id" json:"user_id"` // 用户ID
	RoleID    uint      `gorm:"primaryKey;column:role_id" json:"role_id"` // 角色ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`         // 创建时间
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}
