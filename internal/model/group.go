package model

import (
	"time"

	"gorm.io/gorm"
)

// Group 群组模型
type Group struct {
	ID              uint64         `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	Name            string         `gorm:"type:varchar(64);not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	GroupKey        string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"group_key"` // MinIO桶名
	InviteCode      string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"invite_code"`
	InviteExpiresAt *time.Time     `json:"invite_expires_at"`
	StorageQuota    int64          `gorm:"default:0" json:"storage_quota"` // 存储配额，0表示无限制
	CreatorID       uint64         `gorm:"type:bigint unsigned;not null" json:"creator_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Status          int            `gorm:"type:tinyint;default:1;not null" json:"status"` // 1-正常, 2-禁用, 3-锁定
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	
	Creator User `gorm:"foreignKey:CreatorID" json:"creator"`
}

// TableName 表名
func (Group) TableName() string {
	return "groups"
}

// GroupMember 群组成员模型
type GroupMember struct {
	ID           uint64    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	GroupID      uint64    `gorm:"type:bigint unsigned;not null;index:idx_group_user,priority:1" json:"group_id"`
	UserID       uint64    `gorm:"type:bigint unsigned;not null;index:idx_group_user,priority:2" json:"user_id"`
	Role         string    `gorm:"type:varchar(20);not null" json:"role"` // admin(管理员), member(普通成员)
	JoinedAt     time.Time `gorm:"not null" json:"joined_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastActiveAt *time.Time `json:"last_active_at"`
	
	Group Group `gorm:"foreignKey:GroupID" json:"group"`
	User  User  `gorm:"foreignKey:UserID" json:"user"`
}

// TableName 表名
func (GroupMember) TableName() string {
	return "group_members"
} 