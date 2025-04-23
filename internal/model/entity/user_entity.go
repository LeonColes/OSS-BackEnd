package entity

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint64         `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	Email        string         `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`
	Name         string         `gorm:"type:varchar(64);not null" json:"name"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Avatar       string         `gorm:"type:varchar(255)" json:"avatar"`
	RoleID       uint64         `gorm:"type:bigint unsigned;not null;default:2" json:"role_id"` // 默认为普通用户角色
	IsAdmin      bool           `gorm:"default:false;not null" json:"is_admin"`                 // 是否为系统管理员
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"type:varchar(50)" json:"last_login_ip"`
	Status       int            `gorm:"type:tinyint;default:1;not null" json:"status"` // 1-正常, 2-禁用, 3-锁定
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Role Role `gorm:"foreignKey:RoleID" json:"role"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
