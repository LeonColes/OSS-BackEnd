package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oss-backend/internal/config"
	"oss-backend/internal/controller"
	"oss-backend/internal/model/entity"
	"oss-backend/pkg/minio"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Application 应用程序结构体
type Application struct {
	Config     *config.Config
	DB         *gorm.DB
	Redis      *redis.Client
	Minio      *minio.Client
	Router     *gin.Engine
	Logger     *zap.Logger
	ShutdownCh chan os.Signal
}

// Setup 设置应用程序
func Setup() (*Application, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化日志
	logger := initLogger(cfg)

	// 初始化数据库
	db, err := initDatabase(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化Redis
	redisClient, err := initRedis(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}

	// 初始化MinIO
	minioClient, err := initMinIO(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("初始化MinIO失败: %w", err)
	}

	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin路由引擎
	r := gin.New()
	r.Use(gin.Recovery())

	// 注册路由
	controller.RegisterRoutes(r, db, minioClient)

	// 创建应用程序实例
	app := &Application{
		Config:     cfg,
		DB:         db,
		Redis:      redisClient,
		Minio:      minioClient,
		Router:     r,
		Logger:     logger,
		ShutdownCh: make(chan os.Signal, 1),
	}

	return app, nil
}

// Run 运行应用程序
func (app *Application) Run() error {
	// 设置信号处理
	signal.Notify(app.ShutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	addr := fmt.Sprintf(":%s", app.Config.Server.Port)
	app.Logger.Info("服务器启动", zap.String("地址", addr))

	// 创建服务器
	go func() {
		if err := app.Router.Run(addr); err != nil {
			app.Logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	<-app.ShutdownCh
	app.Logger.Info("正在关闭服务器...")

	// 优雅关闭上下文
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭数据库连接
	sqlDB, err := app.DB.DB()
	if err != nil {
		app.Logger.Error("获取数据库连接失败", zap.Error(err))
	} else {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Error("关闭数据库连接失败", zap.Error(err))
		}
	}

	// 关闭Redis连接
	if err := app.Redis.Close(); err != nil {
		app.Logger.Error("关闭Redis连接失败", zap.Error(err))
	}

	app.Logger.Info("服务器已关闭")
	return nil
}

// Shutdown 关闭应用程序
func (app *Application) Shutdown() {
	app.ShutdownCh <- syscall.SIGTERM
}

// 初始化日志
func initLogger(cfg *config.Config) *zap.Logger {
	var zapLogger *zap.Logger
	var err error

	if cfg.Log.Format == "json" {
		zapLogger, err = zap.NewProduction()
	} else {
		zapLogger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	zap.ReplaceGlobals(zapLogger)
	return zapLogger
}

// 初始化数据库
func initDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	// 设置GORM日志级别
	var gormLogLevel gormlogger.LogLevel
	if cfg.Server.Mode == "production" {
		gormLogLevel = gormlogger.Error
	} else {
		gormLogLevel = gormlogger.Info
	}

	// 配置GORM日志
	gormLogger := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormLogLevel,
			Colorful:      true,
		},
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// 自动迁移数据库表结构
	if err := entity.Setup(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化基础数据
	if err := entity.SeedData(db); err != nil {
		return nil, fmt.Errorf("数据初始化失败: %w", err)
	}

	logger.Info("数据库初始化完成")
	return db, nil
}

// 初始化Redis
func initRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	logger.Info("Redis初始化完成")
	return redisClient, nil
}

// 初始化MinIO
func initMinIO(cfg *config.Config, logger *zap.Logger) (*minio.Client, error) {
	// 初始化MinIO客户端
	minioConfig := minio.Config{
		Endpoint:  cfg.MinIO.Endpoint,
		AccessKey: cfg.MinIO.AccessKey,
		SecretKey: cfg.MinIO.SecretKey,
		UseSSL:    cfg.MinIO.UseSSL,
	}

	minioClient, err := minio.NewClient(minioConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化MinIO客户端失败: %w", err)
	}

	logger.Info("MinIO初始化完成")
	return minioClient, nil
}
