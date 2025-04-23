package model

import (
	"log"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Setup 初始化数据库模型
func Setup(db *gorm.DB) error {
	// 自动迁移表结构
	return db.AutoMigrate(
		&User{},
		&Group{},
		&GroupMember{},
		&Project{},
		&Permission{},
		&File{},
		&FileVersion{},
		&FileShare{},
		&Log{},
		&StorageStat{},
	)
}

// SeedData 初始化基础数据（如果不存在）
func SeedData(db *gorm.DB) error {
	// 检查是否已有用户，如果有则跳过
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}
	
	// 如果已经有用户记录，则不需要初始化
	if count > 0 {
		return nil
	}
	
	// 创建默认管理员账户
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := &User{
		Email:        "admin@example.com",
		Name:         "管理员",
		PasswordHash: string(adminPassword),
		IsAdmin:      true,
		Status:       1, // 正常状态
	}
	
	if err := db.Create(admin).Error; err != nil {
		return err
	}
	
	log.Println("初始化管理员账户成功，邮箱: admin@example.com, 密码: admin123")
	
	// 创建默认群组
	defaultGroup := &Group{
		Name:        "默认群组",
		Description: "系统自动创建的默认群组",
		GroupKey:    "default-group",
		InviteCode:  "default123",
		StorageQuota: 0, // 无限制
		CreatorID:   admin.ID,
		Status:      1, // 正常状态
	}
	
	if err := db.Create(defaultGroup).Error; err != nil {
		return err
	}
	
	// 添加管理员到默认群组
	groupMember := &GroupMember{
		GroupID:  defaultGroup.ID,
		UserID:   admin.ID,
		Role:     "group_owner", // 群组所有者
		JoinedAt: time.Now(),    // 设置加入时间为当前时间
	}
	
	if err := db.Create(groupMember).Error; err != nil {
		return err
	}
	
	// 创建默认项目
	defaultProject := &Project{
		GroupID:     defaultGroup.ID,
		Name:        "默认项目",
		Description: "系统自动创建的默认项目",
		PathPrefix:  "default",
		CreatorID:   admin.ID,
		Status:      1, // 正常状态
	}
	
	if err := db.Create(defaultProject).Error; err != nil {
		return err
	}
	
	// 添加管理员项目权限
	permission := &Permission{
		UserID:    admin.ID,
		ProjectID: defaultProject.ID,
		Role:      "admin", // 项目管理员
		GrantedBy: admin.ID,
	}
	
	if err := db.Create(permission).Error; err != nil {
		return err
	}
	
	log.Println("初始化基础数据成功")
	return nil
} 