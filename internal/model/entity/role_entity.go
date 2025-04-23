package entity

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          uint64         `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	Name        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Group role constants
const (
	GroupRoleOwner  = "owner"  // 拥有者
	GroupRoleAdmin  = "admin"  // 管理员
	GroupRoleMember = "member" // 成员
	GroupRoleGuest  = "guest"  // 访客
)

// Project role constants
const (
	ProjectRoleAdmin  = "admin"  // 管理员
	ProjectRoleEditor = "editor" // 编辑者
	ProjectRoleViewer = "viewer" // 查看者
	ProjectRoleNone   = "none"   // 无权限
)

// TableName 表名
func (Role) TableName() string {
	return "roles"
}
