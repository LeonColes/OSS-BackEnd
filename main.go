package main

import (
	"log"

	"github.com/gin-gonic/gin"

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
	// 初始化应用
	r := gin.Default()

	// 设置路由
	// 实际项目中这里应该传入数据库连接
	routes.SetupRouter(r, nil)

	// 启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
