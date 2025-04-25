package routes

import (
	_ "oss-backend/docs/swagger"        // 统一Swagger文档导入路径
	_ "oss-backend/internal/controller" // 导入控制器包以确保Swagger正确扫描

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"oss-backend/internal/controller"
	"oss-backend/internal/middleware"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
	"oss-backend/pkg/minio"
)

// SetupRouter 设置路由 (接收 Enforcer)
func SetupRouter(r *gin.Engine, db *gorm.DB, enforcer *casbin.Enforcer, minioClient *minio.Client) {
	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	fileRepo := repository.NewFileRepository(db)
	casbinRepo := repository.NewCasbinRepository(db)

	// 创建JWT中间件
	jwtMiddleware := middleware.NewJWTAuthMiddleware()

	// 创建统一的认证授权服务 (需要 Enforcer, 在 main.go 初始化)
	authService := service.NewAuthService(enforcer, roleRepo, userRepo, casbinRepo, db)

	// 创建认证与授权中间件 (传入 Enforcer)
	authMiddleware := middleware.NewAuthMiddleware(authService, userRepo, enforcer)

	// API 路由组
	apiGroup := r.Group("/api/oss")
	{
		// 注册用户相关路由
		registerUserRoutes(apiGroup, userRepo, roleRepo, jwtMiddleware, authMiddleware, authService)

		// 注册群组相关路由
		registerGroupRoutes(apiGroup, userRepo, roleRepo, groupRepo, jwtMiddleware, authMiddleware, authService, minioClient)

		// 注册项目相关路由
		registerProjectRoutes(apiGroup, projectRepo, groupRepo, userRepo, fileRepo, jwtMiddleware, authMiddleware, authService, db, minioClient)

		// 注册文件相关路由
		registerFileRoutes(apiGroup, fileRepo, projectRepo, minioClient, jwtMiddleware, authMiddleware, authService, db)
	}
}

// 注册角色相关路由
func registerRoleRoutes(
	apiGroup *gin.RouterGroup,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
) {
	// 创建依赖
	roleController := controller.NewRoleController(authService)

	// 角色相关路由 - 需要管理员权限
	roleGroup := apiGroup.Group("/role")
	roleGroup.Use(jwtMiddleware.AuthMiddleware())
	roleGroup.Use(authMiddleware.RequireAnyRole("ADMIN", "GROUP_ADMIN"))
	{
		roleGroup.POST("/create", roleController.CreateRole)
		roleGroup.POST("/update", roleController.UpdateRole)
		roleGroup.GET("/delete/:id", roleController.DeleteRole)
		roleGroup.GET("/detail/:id", roleController.GetRoleByID)
		roleGroup.GET("/list", roleController.ListRoles)
	}
}

// 注册用户相关路由
func registerUserRoutes(
	apiGroup *gin.RouterGroup,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
) {
	// 创建依赖
	userService := service.NewUserService(userRepo, roleRepo, authService)
	userController := controller.NewUserController(userService)

	// 用户相关路由
	userGroup := apiGroup.Group("/user")
	{
		// 公共路由，不需要认证
		userGroup.POST("/register", userController.Register)
		userGroup.POST("/login", userController.Login)

		// 认证路由组
		authGroup := userGroup.Group("/")
		authGroup.Use(jwtMiddleware.AuthMiddleware())
		{
			// 基本用户信息 - 需要登录
			authGroup.GET("/info", userController.GetUserInfo)
			authGroup.POST("/update", userController.UpdateUserInfo)
			authGroup.POST("/password", userController.UpdatePassword)

			// 用户管理 - 需要管理员权限
			adminGroup := authGroup.Group("/")
			adminGroup.Use(authMiddleware.RequireAnyRole("GROUP_ADMIN"))
			{
				adminGroup.GET("/list", userController.ListUsers)
				adminGroup.GET("/status/:id", userController.UpdateUserStatus)

				// 用户角色管理
				adminGroup.GET("/roles/:id", userController.GetUserRoles)
				adminGroup.POST("/roles/:id", userController.AssignRoles)
				adminGroup.POST("/roles/:id/remove", userController.RemoveRoles)
			}
		}
	}
}

// 注册群组相关路由
func registerGroupRoutes(
	apiGroup *gin.RouterGroup,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	groupRepo repository.GroupRepository,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
	minioClient *minio.Client, // 添加MinIO客户端参数
) {
	// 创建依赖
	groupService := service.NewGroupService(groupRepo, userRepo, roleRepo, authService, minioClient)
	groupController := controller.NewGroupController(groupService)

	// 群组相关路由
	groupGroup := apiGroup.Group("/group")
	groupGroup.Use(jwtMiddleware.AuthMiddleware()) // 所有群组操作都需要登录
	{
		// 群组管理
		groupGroup.POST("/create", groupController.CreateGroup)
		groupGroup.POST("/update", groupController.UpdateGroup)
		groupGroup.GET("/detail/:id", groupController.GetGroupByID)
		groupGroup.GET("/list", groupController.ListGroups)
		groupGroup.GET("/user", groupController.GetUserGroups)
		groupGroup.POST("/join", groupController.JoinGroup)
		groupGroup.POST("/invite", groupController.GenerateInviteCode)

		// 成员管理 - 需要群组管理员权限
		memberGroup := groupGroup.Group("/member")
		memberGroup.Use(authMiddleware.RequireAdmin())
		{
			memberGroup.GET("/add/:id", groupController.AddMember)
			memberGroup.POST("/role/:id", groupController.UpdateMemberRole)
			memberGroup.GET("/remove/:id", groupController.RemoveMember)
			memberGroup.GET("/list/:id", groupController.ListMembers)
		}
	}
}

// 注册项目相关路由
func registerProjectRoutes(
	apiGroup *gin.RouterGroup,
	projectRepo repository.ProjectRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	fileRepo repository.FileRepository,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
	db *gorm.DB,
	minioClient *minio.Client,
) {
	// 初始化项目仓库和服务
	projectService := service.NewProjectService(
		projectRepo,
		groupRepo,
		userRepo,
		authService,
		db,
		minioClient,
	)
	projectController := controller.NewProjectController(projectService)

	// 定义中间件辅助函数
	getProjectGroupID := func(c *gin.Context) (string, error) {
		return middleware.GetGroupIDFromParam(c)
	}

	// 项目相关路由
	projectGroup := apiGroup.Group("/project")
	projectGroup.Use(jwtMiddleware.AuthMiddleware())
	{
		// 项目管理
		projectGroup.POST("/create", authMiddleware.Authorize("projects", "create", getProjectGroupID), projectController.CreateProject)
		projectGroup.POST("/update", authMiddleware.Authorize("projects", "update", getProjectGroupID), projectController.UpdateProject)
		projectGroup.GET("/detail/:id", authMiddleware.Authorize("projects", "read", getProjectGroupID), projectController.GetProjectByID)
		projectGroup.GET("/delete/:id", authMiddleware.Authorize("projects", "delete", getProjectGroupID), projectController.DeleteProject)
		projectGroup.GET("/list", authMiddleware.Authorize("projects", "read", getProjectGroupID), projectController.ListProjects)
		projectGroup.GET("/user", projectController.GetUserProjects)

		// 项目成员管理 - 需要群组管理员权限
		memberGroup := projectGroup.Group("/member")
		memberGroup.Use(authMiddleware.RequireAdmin())
		{
			memberGroup.POST("/add", projectController.SetPermission)
			memberGroup.POST("/remove", projectController.RemovePermission)
			memberGroup.GET("/list/:id", projectController.ListProjectUsers)
		}
	}
}

// 注册文件相关路由
func registerFileRoutes(
	apiGroup *gin.RouterGroup,
	fileRepo repository.FileRepository,
	projectRepo repository.ProjectRepository,
	minioClient *minio.Client,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
	db *gorm.DB,
) {
	// 创建文件服务
	fileService := service.NewFileService(fileRepo, projectRepo, minioClient, authService, db)

	// 创建文件控制器
	fileController := controller.NewFileController(fileService, nil, authService)

	// 定义文件中间件辅助函数
	getFileGroupID := func(c *gin.Context) (string, error) {
		return middleware.GetGroupIDFromParam(c)
	}

	// 文件相关路由
	fileGroup := apiGroup.Group("/file")
	fileGroup.Use(jwtMiddleware.AuthMiddleware())
	{
		// 文件管理路由
		fileGroup.POST("/upload", authMiddleware.Authorize("files", "create", getFileGroupID), fileController.Upload)
		fileGroup.GET("/download/:id", authMiddleware.Authorize("files", "read", getFileGroupID), fileController.Download)
		fileGroup.GET("/public-url/:id", authMiddleware.Authorize("files", "read", getFileGroupID), fileController.GetPublicURL)
		fileGroup.GET("/list", authMiddleware.Authorize("files", "read", getFileGroupID), fileController.ListFiles)
		fileGroup.POST("/folder", authMiddleware.Authorize("files", "create", getFileGroupID), fileController.CreateFolder)
		fileGroup.GET("/delete/:id", authMiddleware.Authorize("files", "delete", getFileGroupID), fileController.DeleteFile)
		fileGroup.GET("/versions/:id", authMiddleware.Authorize("files", "read", getFileGroupID), fileController.GetFileVersions)
	}

	// 文件分享相关路由
	shareGroup := apiGroup.Group("/share")
	{
		// 创建分享需要认证
		shareGroup.POST("", jwtMiddleware.AuthMiddleware(), fileController.CreateShare)

		// 获取分享信息与下载分享文件不需要认证
		shareGroup.GET("/:code", fileController.GetShareInfo)
		shareGroup.POST("/download", fileController.DownloadSharedFile)
	}
}
