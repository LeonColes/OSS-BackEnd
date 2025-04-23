package entity

import (
	"time"
)

// Role 角色实体，对应数据库中的角色表
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;not null;uniqueIndex" json:"name"` // 角色名称
	Description string    `gorm:"size:200" json:"description"`              // 角色描述
	Code        string    `gorm:"size:50;not null;uniqueIndex" json:"code"` // 角色编码，用于权限控制
	Status      int       `gorm:"default:1" json:"status"`                  // 状态：1-启用，0-禁用
	IsSystem    bool      `gorm:"default:false" json:"is_system"`           // 是否为系统角色，系统角色不可删除
	CreatedBy   uint      `json:"created_by"`                               // 创建者ID
	UpdatedBy   uint      `json:"updated_by"`                               // 更新者ID
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`         // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`         // 更新时间
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// 预定义系统角色
const (
	RoleGroupAdmin   = "GROUP_ADMIN"   // 群组管理员
	RoleProjectAdmin = "PROJECT_ADMIN" // 项目管理员
	RoleMember       = "MEMBER"        // 普通成员
)
