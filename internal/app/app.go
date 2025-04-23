package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/leoncoles/oss-backend/internal/api"
	"github.com/leoncoles/oss-backend/internal/config"
	
	// 导入swagger文档
	_ "github.com/leoncoles/oss-backend/docs/swagger"
	
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 应用程序结构
type Application struct {
	config      *config.Config
	db          *gorm.DB
	redisClient *redis.Client
	minioClient *minio.Client
	logger      *zap.Logger
	router      *gin.Engine
}

// NewApplication 创建应用程序实例
func NewApplication(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, minioClient *minio.Client, logger *zap.Logger) *Application {
	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	return &Application{
		config:      cfg,
		db:          db,
		redisClient: redisClient,
		minioClient: minioClient,
		logger:      logger,
		router:      router,
	}
}

// Run 运行应用程序
func (a *Application) Run() error {
	// 初始化路由
	a.setupRoutes()

	// 启动HTTP服务器
	addr := fmt.Sprintf(":%s", a.config.Server.Port)
	a.logger.Info("Starting server", zap.String("address", addr))
	return a.router.Run(addr)
}

// setupRoutes 设置路由
func (a *Application) setupRoutes() {
	// 注册API路由
	api.RegisterRoutes(a.router, a.db)
	
	// 添加静态文件目录，使swagger.json可直接访问
	a.router.Static("/swagger-json", "./docs/swagger")
	
	// 添加Swagger文档路由
	url := ginSwagger.URL("/swagger-json/swagger.json")
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	
	// 添加根路径重定向到Swagger
	a.router.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})
} 