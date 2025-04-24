package routes

import (
	_ "oss-backend/docs/swagger" // 统一Swagger文档导入路径

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"oss-backend/internal/controller"
	"oss-backend/internal/middleware"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
	"oss-backend/pkg/minio"
)

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine, db interface{}) {
	// 转换数据库连接
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("数据库连接类型错误")
	}

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 创建中间件
	jwtMiddleware := middleware.NewJWTAuthMiddleware()

	// 创建仓库
	userRepo := repository.NewUserRepository(gormDB)
	roleRepo := repository.NewRoleRepository(gormDB)
	groupRepo := repository.NewGroupRepository(gormDB)
	casbinRepo := repository.NewCasbinRepository(gormDB)

	// 初始化Casbin执行器
	enforcer, err := casbin.NewEnforcer("configs/rbac_model.conf", "configs/policy.csv")
	if err != nil {
		panic("初始化Casbin执行器失败: " + err.Error())
	}

	// 创建统一的认证授权服务
	authSvc := service.NewAuthService(enforcer, roleRepo, userRepo, casbinRepo, gormDB)

	// 初始化RBAC权限
	err = authSvc.InitializeRBAC()
	if err != nil {
		panic("初始化RBAC权限失败: " + err.Error())
	}

	// 创建认证与授权中间件
	authMiddleware := middleware.NewAuthMiddleware(authSvc, userRepo, enforcer)

	// 设置 API 前缀
	apiGroup := r.Group("/api/oss")

	// 注册角色相关路由
	registerRoleRoutes(apiGroup, jwtMiddleware, authMiddleware, authSvc)

	// 注册用户相关路由
	registerUserRoutes(apiGroup, gormDB, jwtMiddleware, authMiddleware, authSvc)

	// 注册群组相关路由
	registerGroupRoutes(apiGroup, userRepo, roleRepo, groupRepo, jwtMiddleware, authMiddleware)

	// 注册项目相关路由
	registerProjectRoutes(apiGroup, gormDB, jwtMiddleware, authSvc, authMiddleware)

	// 注册文件相关路由
	registerFileRoutes(apiGroup, gormDB, jwtMiddleware, authMiddleware, authSvc)
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
	db *gorm.DB,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
) {
	// 创建依赖
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
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
) {
	// 创建依赖
	groupService := service.NewGroupService(groupRepo, userRepo, roleRepo)
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
	db *gorm.DB,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authService service.AuthService,
	authMiddleware *middleware.AuthMiddleware,
) {
	// 初始化项目仓库和服务
	projectRepo := repository.NewProjectRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	userRepo := repository.NewUserRepository(db)

	projectService := service.NewProjectService(
		projectRepo,
		groupRepo,
		userRepo,
		authService,
		db,
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
	db *gorm.DB,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	authService service.AuthService,
) {
	// 创建仓库
	fileRepo := repository.NewFileRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	// 从配置获取MinIO参数
	minioConfig := minio.Config{
		Endpoint:  viper.GetString("minio.endpoint"),
		AccessKey: viper.GetString("minio.access_key"),
		SecretKey: viper.GetString("minio.secret_key"),
		UseSSL:    viper.GetBool("minio.use_ssl"),
	}

	// 如果配置为空，使用默认值
	if minioConfig.Endpoint == "" {
		minioConfig.Endpoint = "localhost:9000"
		minioConfig.AccessKey = "minioadmin"
		minioConfig.SecretKey = "minioadmin"
		minioConfig.UseSSL = false
	}

	// 创建MinIO客户端
	minioClient, err := minio.NewClient(minioConfig)
	if err != nil {
		panic("初始化MinIO客户端失败: " + err.Error())
	}

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
