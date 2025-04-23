package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leoncoles/oss-backend/internal/app"
	"github.com/leoncoles/oss-backend/internal/config"
	"github.com/leoncoles/oss-backend/internal/model"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	zapLogger := initLogger(cfg)
	defer zapLogger.Sync()
	
	// 初始化数据库
	db := initDatabase(cfg, zapLogger)
	
	// 初始化Redis
	redisClient := initRedis(cfg, zapLogger)
	
	// 初始化MinIO
	minioClient := initMinIO(cfg, zapLogger)

	// 创建应用
	application := app.NewApplication(cfg, db, redisClient, minioClient, zapLogger)
	
	// 在后台运行应用
	go func() {
		if err := application.Run(); err != nil {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	zapLogger.Info("Shutting down server...")
	
	// 优雅关闭上下文
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 关闭数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Error("Failed to get database connection", zap.Error(err))
	} else {
		sqlDB.Close()
	}
	
	// 关闭Redis连接
	if err := redisClient.Close(); err != nil {
		zapLogger.Error("Failed to close Redis connection", zap.Error(err))
	}
	
	zapLogger.Info("Server exiting")
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
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	
	zap.ReplaceGlobals(zapLogger)
	return zapLogger
}

// 初始化数据库
func initDatabase(cfg *config.Config, zapLogger *zap.Logger) *gorm.DB {
	// 设置GORM日志级别
	gormLogLevel := logger.Info
	if cfg.Server.Mode == "production" {
		gormLogLevel = logger.Error
	}
	
	// 配置GORM日志
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
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
		zapLogger.Fatal("Failed to connect database", zap.Error(err))
	}
	
	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("Failed to get database connection", zap.Error(err))
	}
	
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	
	// 自动迁移数据库表结构
	if err := model.Setup(db); err != nil {
		zapLogger.Fatal("Failed to migrate database", zap.Error(err))
	}
	
	// 初始化基础数据
	if err := model.SeedData(db); err != nil {
		zapLogger.Fatal("Failed to seed database", zap.Error(err))
	}
	
	zapLogger.Info("Database initialized successfully")
	return db
}

// 初始化Redis
func initRedis(cfg *config.Config, zapLogger *zap.Logger) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		zapLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	
	zapLogger.Info("Redis initialized successfully")
	return redisClient
}

// 初始化MinIO
func initMinIO(cfg *config.Config, zapLogger *zap.Logger) *minio.Client {
	// 初始化MinIO客户端
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	
	if err != nil {
		zapLogger.Fatal("Failed to initialize MinIO client", zap.Error(err))
	}
	
	zapLogger.Info("MinIO initialized successfully")
	return minioClient
}

