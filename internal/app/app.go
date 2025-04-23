package app

import (
	"github.com/gin-gonic/gin"
	"github.com/leoncoles/oss-backend/internal/api"
	"github.com/leoncoles/oss-backend/internal/config"
	"github.com/leoncoles/oss-backend/internal/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 应用结构体
type Application struct {
	Config      *config.Config
	DB          *gorm.DB
	Redis       *redis.Client
	MinIO       *minio.Client
	Logger      *zap.Logger
	Router      *gin.Engine
}

// NewApplication 创建应用实例
func NewApplication(
	cfg *config.Config,
	db *gorm.DB,
	redis *redis.Client,
	minioClient *minio.Client,
	logger *zap.Logger,
) *Application {
	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// 初始化路由
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	// 开发环境暂时不需要跨域中间件
	// router.Use(middleware.Cors())
	
	return &Application{
		Config: cfg,
		DB:     db,
		Redis:  redis,
		MinIO:  minioClient,
		Logger: logger,
		Router: router,
	}
}

// SetupRoutes 设置路由
func (app *Application) SetupRoutes() {
	api.RegisterRoutes(app.Router, app.DB)
}

// Run 运行应用
func (app *Application) Run() error {
	// 设置路由
	app.SetupRoutes()

	// 运行HTTP服务
	addr := ":" + app.Config.Server.Port
	app.Logger.Info("Starting server", zap.String("address", addr))
	return app.Router.Run(addr)
} 