package entity

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Setup 自动迁移所有表结构
func Setup(db *gorm.DB) error {
	log.Println("开始数据库迁移...")

	// 自动迁移表结构
	err := db.AutoMigrate(
		&User{},
		&Role{},
		&Group{},
		&GroupMember{},
		&Project{},
		&File{},
		&FileShare{},
		&OperationLog{},
		&AccessLog{},
	)

	if err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Println("数据库迁移完成")
	return nil
}

// SeedData 初始化基础数据
func SeedData(db *gorm.DB) error {
	log.Println("开始初始化基础数据...")

	// 初始化角色
	if err := seedRoles(db); err != nil {
		return err
	}

	// 初始化管理员账户
	if err := seedAdminUser(db); err != nil {
		return err
	}

	log.Println("基础数据初始化完成")
	return nil
}

// seedRoles 初始化角色数据
func seedRoles(db *gorm.DB) error {
	roles := []Role{
		{Name: "admin", Description: "管理员，拥有所有权限"},
		{Name: "user", Description: "普通用户"},
	}

	// 检查是否已存在角色数据
	var count int64
	db.Model(&Role{}).Count(&count)
	if count > 0 {
		return nil // 已有角色数据，跳过
	}

	// 创建角色
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return fmt.Errorf("创建角色失败: %w", err)
		}
	}

	return nil
}

// seedAdminUser 初始化管理员账户
func seedAdminUser(db *gorm.DB) error {
	// 检查是否已存在管理员账户
	var count int64
	db.Model(&User{}).Where("email = ?", "admin@example.com").Count(&count)
	if count > 0 {
		return nil // 已有管理员账户，跳过
	}

	// 创建密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	// 获取管理员角色ID
	var adminRole Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("获取管理员角色失败: %w", err)
	}

	// 创建管理员用户
	admin := User{
		Name:         "管理员",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		RoleID:       adminRole.ID,
		IsAdmin:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("创建管理员账户失败: %w", err)
	}

	return nil
}
