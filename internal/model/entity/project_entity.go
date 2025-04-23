package entity

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID          uint64         `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	GroupID     uint64         `gorm:"type:bigint unsigned;not null;index" json:"group_id"`
	Name        string         `gorm:"type:varchar(64);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	PathPrefix  string         `gorm:"type:varchar(128);not null" json:"path_prefix"`
	CreatorID   uint64         `gorm:"type:bigint unsigned;not null" json:"creator_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Status      int            `gorm:"type:tinyint;default:1;not null" json:"status"` // 1-正常, 2-归档, 3-删除
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Group   Group `gorm:"foreignKey:GroupID" json:"group"`
	Creator User  `gorm:"foreignKey:CreatorID" json:"creator"`
}

// TableName 表名
func (Project) TableName() string {
	return "projects"
}

// Permission 项目权限模型
type Permission struct {
	ID        uint64     `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	UserID    uint64     `gorm:"type:bigint unsigned;not null;index:idx_project_user,priority:2" json:"user_id"`
	ProjectID uint64     `gorm:"type:bigint unsigned;not null;index:idx_project_user,priority:1" json:"project_id"`
	Role      string     `gorm:"type:varchar(20);not null" json:"role"` // admin, editor, viewer
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpireAt  *time.Time `json:"expire_at"`
	GrantedBy uint64     `gorm:"type:bigint unsigned;not null" json:"granted_by"`

	User    User    `gorm:"foreignKey:UserID" json:"user"`
	Project Project `gorm:"foreignKey:ProjectID" json:"project"`
	Granter User    `gorm:"foreignKey:GrantedBy" json:"granter"`
}

// TableName 表名
func (Permission) TableName() string {
	return "permissions"
}

// 角色常量在 roles.go 中定义
