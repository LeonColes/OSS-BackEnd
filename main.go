package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// 导入Swagger文档
	_ "oss-backend/docs/swagger"
	// 导入所有控制器以确保扫描到API注释
	_ "oss-backend/internal/controller"

	"oss-backend/internal/model/entity"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
	"oss-backend/pkg/minio"
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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 在请求头中添加Bearer令牌，格式为: Bearer {token}

// @Security BearerAuth

// @Tag.name 系统管理员API
// @Tag.description 需要系统管理员权限的接口

// @Tag.name 用户模块
// @Tag.description 用户相关的接口

// @Tag.name 角色管理
// @Tag.description 角色相关的接口

// @Tag.name 群组管理
// @Tag.description 群组相关的接口

// @Tag.name 项目管理
// @Tag.description 项目相关的接口

// @Tag.name 文件管理
// @Tag.description 文件相关的接口

// @Tag.name 文件分享
// @Tag.description 文件分享相关的接口

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

	// 初始化MinIO客户端
	minioConfig := minio.Config{
		Endpoint:  viper.GetString("minio.endpoint"),
		AccessKey: viper.GetString("minio.access_key"),
		SecretKey: viper.GetString("minio.secret_key"),
		UseSSL:    viper.GetBool("minio.use_ssl"),
	}

	minioClient, err := minio.NewClient(minioConfig)
	if err != nil {
		log.Fatalf("初始化MinIO客户端失败: %v", err)
	}

	// 初始化 Casbin Enforcer
	enforcer, err := initCasbin(db)
	if err != nil {
		log.Fatalf("初始化 Casbin 失败: %v", err)
	}

	// 初始化角色和管理员用户 (传入 Enforcer)
	if err := initRolesAndAdmin(db, enforcer); err != nil {
		log.Printf("初始化角色和管理员用户失败: %v", err)
	}

	// 初始化应用
	r := gin.Default()

	// 设置路由 (需要将 Enforcer 传递下去，或者通过依赖注入)
	routes.SetupRouter(r, db, enforcer, minioClient)

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
		dsn = "root:password@tcp(127.0.0.1:3306)/oss?charset=utf8mb4&parseTime=True&loc=Local"
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
		&entity.ProjectMember{},
		&entity.Permission{},
		&entity.File{},
		&entity.FileVersion{},
		&entity.FileShare{},
		&entity.Group{},
		&entity.GroupMember{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 初始化 Casbin Enforcer
func initCasbin(db *gorm.DB) (*casbin.Enforcer, error) {
	// 1. 创建 Gorm Adapter
	// Gorm Adapter 默认使用的表名是 'casbin_rule'，如果表名不同需要配置
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("创建 casbin adapter 失败: %w", err)
	}

	// 2. 创建 Enforcer
	// 确认模型文件路径正确
	enforcer, err := casbin.NewEnforcer("configs/rbac_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("创建 casbin enforcer 失败: %w", err)
	}

	// 3. 加载策略 (Adapter 通常会自动加载，但显式调用更保险)
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("加载 casbin policy 失败: %w", err)
	}

	log.Println("Casbin Enforcer 初始化成功")
	return enforcer, nil
}

// 初始化角色和管理员用户 (接收 Enforcer)
func initRolesAndAdmin(db *gorm.DB, enforcer *casbin.Enforcer) error {
	ctx := context.Background()

	// 初始化仓库
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 初始化基础系统角色
	initSystemRoles(ctx, roleRepo)

	// 初始化服务 (传入 Enforcer)
	casbinRepo := repository.NewCasbinRepository(db)
	authService := service.NewAuthService(enforcer, roleRepo, userRepo, casbinRepo, db)
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
