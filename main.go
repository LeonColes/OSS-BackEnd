package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
	"oss-backend/routes"
)

// @title OSS-Backend API
// @version 1.0
// @description 企业级对象存储系统后端 API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	// 初始化配置
	if err := initConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化角色和管理员用户
	if err := initRolesAndAdmin(db); err != nil {
		log.Printf("初始化角色和管理员用户失败: %v", err)
	}

	// 初始化应用
	r := gin.Default()

	// 设置路由
	routes.SetupRouter(r, db)

	// 读取服务器端口配置
	port := viper.GetInt("server.port")
	if port == 0 {
		port = 8081 // 默认端口
	}

	// 启动服务
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}

// 初始化配置
func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	return viper.ReadInConfig()
}

// 初始化数据库
func initDB() (*gorm.DB, error) {
	// 从配置文件读取数据库连接信息
	dsn := viper.GetString("database.dsn")
	if dsn == "" {
		// 使用默认值
		dsn = "root:password@tcp(127.0.0.1:3306)/oss_backend?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 先连接到MySQL服务器，不指定数据库
	dsnParts := strings.Split(dsn, "/")
	if len(dsnParts) < 2 {
		return nil, fmt.Errorf("DSN格式错误: %s", dsn)
	}

	// 提取数据库名
	dbNameParts := strings.Split(dsnParts[len(dsnParts)-1], "?")
	dbName := dbNameParts[0]

	// 构建不包含数据库名的DSN
	rootDSN := strings.Join(dsnParts[:len(dsnParts)-1], "/") + "/"
	if len(dbNameParts) > 1 {
		rootDSN += "?" + dbNameParts[1]
	}

	// 连接到MySQL
	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %w", err)
	}

	// 创建数据库
	err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)).Error
	if err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	// 连接到指定的数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.UserRole{},
		&entity.Log{},
		&entity.Project{},
		&entity.File{},
		&entity.Group{},
		&entity.GroupMember{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 初始化角色和管理员用户
func initRolesAndAdmin(db *gorm.DB) error {
	ctx := context.Background()

	// 初始化仓库
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 初始化基础系统角色
	initSystemRoles(ctx, roleRepo)

	// 初始化服务
	casbinRepo := repository.NewCasbinRepository(db)
	authService := service.NewAuthService(nil, roleRepo, userRepo, casbinRepo, db)
	userService := service.NewUserService(userRepo, roleRepo, authService)

	// 初始化系统管理员用户
	return userService.InitAdminUser(ctx)
}

// 初始化系统角色
func initSystemRoles(ctx context.Context, roleRepo repository.RoleRepository) {
	// 预定义的系统角色
	systemRoles := []entity.Role{
		{
			Name:        "系统管理员",
			Description: "拥有系统的最高权限",
			Code:        entity.RoleAdmin,
			Status:      1,
			IsSystem:    true,
		},
		{
			Name:        "群组管理员",
			Description: "拥有群组的管理权限",
			Code:        entity.RoleGroupAdmin,
			Status:      1,
			IsSystem:    true,
		},
		{
			Name:        "普通成员",
			Description: "普通用户，具有基本的访问权限",
			Code:        entity.RoleMember,
			Status:      1,
			IsSystem:    true,
		},
	}

	// 创建角色
	for _, role := range systemRoles {
		existRole, _ := roleRepo.GetByCode(ctx, role.Code)
		if existRole == nil {
			_ = roleRepo.Create(ctx, &role)
		}
	}
}
