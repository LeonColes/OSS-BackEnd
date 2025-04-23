package api

import (
	"github.com/gin-gonic/gin"
	"github.com/leoncoles/oss-backend/internal/api/auth"
	"github.com/leoncoles/oss-backend/internal/middleware"
	"github.com/leoncoles/oss-backend/internal/repository"
	"github.com/leoncoles/oss-backend/internal/service"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有API路由
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	
	// 初始化服务
	authService := service.NewAuthService(userRepo)
	
	// 初始化控制器
	authController := auth.NewController(authService)
	
	// API根路径
	api := r.Group("/api/oss")

	// 公开路由 - 不需要认证
	public := api.Group("")
	registerPublicRoutes(public, authController)

	// 受保护路由 - 需要认证
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	registerProtectedRoutes(protected, authController)
}

// registerPublicRoutes 注册公开路由
func registerPublicRoutes(r *gin.RouterGroup, authController *auth.Controller) {
	// 认证相关路由
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/refresh", authController.RefreshToken)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// registerProtectedRoutes 注册受保护路由
func registerProtectedRoutes(r *gin.RouterGroup, authController *auth.Controller) {
	// 用户相关路由
	user := r.Group("/user")
	{
		user.GET("/info", authController.GetUserInfo)
		user.POST("/logout", authController.Logout)
		user.PUT("/password", authController.ChangePassword)
	}

	// 群组相关路由
	group := r.Group("/group")
	{
		group.POST("/create", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "创建群组功能待实现"})
		})
		group.GET("/list", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取群组列表功能待实现"})
		})
		group.GET("/members", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取群组成员功能待实现"})
		})
	}

	// 项目相关路由
	project := r.Group("/project")
	{
		project.POST("/create", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "创建项目功能待实现"})
		})
		project.GET("/list", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取项目列表功能待实现"})
		})
	}

	// 文件相关路由
	file := r.Group("/file")
	{
		file.POST("/instant-check", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "秒传检查功能待实现"})
		})
		file.POST("/upload-token", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取上传Token功能待实现"})
		})
		file.POST("/upload-confirm", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "确认上传功能待实现"})
		})
		file.GET("/list", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取文件列表功能待实现"})
		})
		file.GET("/download-token", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取下载Token功能待实现"})
		})
	}
} 