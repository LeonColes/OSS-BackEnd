package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"oss-backend/internal/model/entity"
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
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化应用
	r := gin.Default()

	// 设置路由
	routes.SetupRouter(r, db)

	// 启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}

// 初始化数据库
func initDB() (*gorm.DB, error) {
	// 使用MySQL数据库
	dsn := "root:password@tcp(127.0.0.1:3306)/oss_backend?charset=utf8mb4&parseTime=True&loc=Local"
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
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
