package controller

import (
	"oss-backend/internal/middleware"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
	"oss-backend/pkg/minio"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有API路由
func RegisterRoutes(router *gin.Engine, db *gorm.DB, minioClient *minio.Client) {
	// 创建存储库
	userRepo := repository.NewUserRepository(db)
	fileRepo := repository.NewFileRepository(db)

	// 创建服务
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(db)
	groupService := service.NewGroupService(db)
	projectService := service.NewProjectService(db)
	fileService := service.NewFileService(fileRepo, minioClient)

	// 创建控制器实例
	authController := NewAuthController(authService, userService)
	userController := NewUserController(userService)
	groupController := NewGroupController(groupService)
	projectController := NewProjectController(projectService)
	fileController := NewFileController(fileService)

	// 设置中间件
	jwtMiddleware := middleware.NewJWTMiddleware(authService)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 公开的API
	router.GET("/swagger/*any", middleware.Swagger())

	// 授权API
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/refresh", authController.RefreshToken)
		authGroup.POST("/register", authController.Register)
	}

	// 需要认证的API
	apiGroup := router.Group("/api")
	apiGroup.Use(jwtMiddleware.AuthMiddleware())
	{
		// 用户API
		userGroup := apiGroup.Group("/user")
		{
			userGroup.GET("/info", userController.GetUserInfo)
			userGroup.PUT("/info", userController.UpdateUserInfo)
			userGroup.PUT("/password", userController.UpdatePassword)
		}

		// 分组API
		groupRoutes := apiGroup.Group("/group")
		{
			groupRoutes.POST("", groupController.CreateGroup)
			groupRoutes.GET("/:id", groupController.GetGroup)
			groupRoutes.PUT("/:id", groupController.UpdateGroup)
			groupRoutes.DELETE("/:id", groupController.DeleteGroup)
			groupRoutes.GET("", groupController.ListGroups)
			groupRoutes.GET("/:id/members", groupController.GetGroupMembers)
			groupRoutes.POST("/:id/members", groupController.AddGroupMember)
			groupRoutes.DELETE("/:id/members/:user_id", groupController.RemoveGroupMember)
			groupRoutes.POST("/:id/members/:user_id/role", groupController.UpdateMemberRole)
		}

		// 项目API
		projectRoutes := apiGroup.Group("/project")
		{
			projectRoutes.POST("", projectController.CreateProject)
			projectRoutes.GET("/:id", projectController.GetProject)
			projectRoutes.PUT("/:id", projectController.UpdateProject)
			projectRoutes.DELETE("/:id", projectController.DeleteProject)
			projectRoutes.GET("", projectController.ListProjects)
			projectRoutes.GET("/:id/members", projectController.GetProjectMembers)
			projectRoutes.POST("/:id/members", projectController.AddProjectMember)
			projectRoutes.DELETE("/:id/members/:user_id", projectController.RemoveProjectMember)
			projectRoutes.POST("/:id/members/:user_id/role", projectController.UpdateMemberRole)
		}

		// 文件API
		fileRoutes := apiGroup.Group("/file")
		{
			fileRoutes.POST("/upload", fileController.UploadFile)
			fileRoutes.POST("/folder", fileController.CreateFolder)
			fileRoutes.GET("/:id", fileController.GetFile)
			fileRoutes.GET("/list", fileController.ListFiles)
			fileRoutes.GET("/:id/download", fileController.DownloadFile)
			fileRoutes.POST("/:id/rename", fileController.RenameFile)
			fileRoutes.POST("/:id/move", fileController.MoveFile)
			fileRoutes.DELETE("/:id", fileController.DeleteFile)
			fileRoutes.POST("/share", fileController.CreateShare)
		}
	}

	// 文件分享API（无需认证）
	shareRoutes := router.Group("/share")
	{
		shareRoutes.GET("/:code", fileController.GetShareFile)
		shareRoutes.GET("/:code/download", fileController.DownloadShareFile)
	}
}

// InitRoutes 初始化路由
func InitRoutes(router *gin.Engine, fileRepo repository.FileRepository, minioClient *minio.Client) {
	// 创建服务
	fileService := service.NewFileService(fileRepo, minioClient)

	// 创建控制器
	fileController := NewFileController(fileService)

	// 创建JWT中间件
	jwtMiddleware := middleware.NewJWTAuthMiddleware()

	// 公开API
	publicAPI := router.Group("/api/v1")
	{
		// 文件分享相关
		publicAPI.GET("/share/:code", fileController.GetShareFile)
		publicAPI.GET("/share/:code/download", fileController.DownloadShareFile)
	}

	// 需要授权的API
	privateAPI := router.Group("/api/v1")
	privateAPI.Use(jwtMiddleware.AuthMiddleware())
	{
		// 文件操作
		fileAPI := privateAPI.Group("/file")
		{
			fileAPI.POST("/upload", fileController.UploadFile)
			fileAPI.POST("/folder", fileController.CreateFolder)
			fileAPI.GET("/:id", fileController.GetFile)
			fileAPI.GET("/list", fileController.ListFiles)
			fileAPI.GET("/:id/download", fileController.DownloadFile)
			fileAPI.POST("/:id/rename", fileController.RenameFile)
			fileAPI.POST("/:id/move", fileController.MoveFile)
			fileAPI.DELETE("/:id", fileController.DeleteFile)
			fileAPI.POST("/share", fileController.CreateShare)
		}
	}
}
