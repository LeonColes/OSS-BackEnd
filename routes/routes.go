package routes

import (
	"context"

	_ "oss-backend/docs/swagger" // 统一Swagger文档导入路径

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"oss-backend/internal/controller"
	"oss-backend/internal/middleware"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
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

	// 创建角色中间件
	roleMiddleware := middleware.NewRoleAuthMiddleware(userRepo)

	// 创建Casbin服务
	casbinSvc, err := service.NewCasbinService(gormDB, userRepo, roleRepo, groupRepo)
	if err != nil {
		panic("创建Casbin服务失败: " + err.Error())
	}

	// 设置 API 前缀
	apiGroup := r.Group("/api/oss")

	// 注册角色相关路由
	registerRoleRoutes(apiGroup, gormDB, jwtMiddleware, roleMiddleware)

	// 注册用户相关路由
	registerUserRoutes(apiGroup, gormDB, jwtMiddleware, roleMiddleware)

	// 注册群组相关路由
	registerGroupRoutes(apiGroup, userRepo, roleRepo, groupRepo, jwtMiddleware)

	// 注册项目相关路由
	registerProjectRoutes(apiGroup, gormDB, jwtMiddleware, casbinSvc)

	// 注册文件相关路由 (空壳)
	registerFileRoutes(apiGroup, jwtMiddleware)
}

// 注册角色相关路由
func registerRoleRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, jwtMiddleware *middleware.JWTAuthMiddleware, roleMiddleware *middleware.RoleAuthMiddleware) {
	// 创建依赖
	roleRepo := repository.NewRoleRepository(db)
	roleService := service.NewRoleService(roleRepo)
	roleController := controller.NewRoleController(roleService)

	// 初始化系统角色
	err := roleService.InitSystemRoles(context.Background())
	if err != nil {
		panic("初始化系统角色失败: " + err.Error())
	}

	// 角色相关路由 - 需要管理员权限
	roleGroup := apiGroup.Group("/role")
	roleGroup.Use(jwtMiddleware.AuthMiddleware())
	roleGroup.Use(roleMiddleware.RequireAnyRole("admin", "super_admin"))
	{
		roleGroup.POST("/create", roleController.CreateRole)
		roleGroup.POST("/update", roleController.UpdateRole)
		roleGroup.GET("/delete/:id", roleController.DeleteRole)
		roleGroup.GET("/detail/:id", roleController.GetRoleByID)
		roleGroup.GET("/list", roleController.ListRoles)
	}
}

// 注册用户相关路由
func registerUserRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, jwtMiddleware *middleware.JWTAuthMiddleware, roleMiddleware *middleware.RoleAuthMiddleware) {
	// 创建依赖
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userService := service.NewUserService(userRepo, roleRepo)
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
			adminGroup.Use(roleMiddleware.RequireAnyRole("admin", "super_admin"))
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
func registerGroupRoutes(apiGroup *gin.RouterGroup, userRepo repository.UserRepository, roleRepo repository.RoleRepository, groupRepo repository.GroupRepository, jwtMiddleware *middleware.JWTAuthMiddleware) {
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

		// 成员管理
		groupGroup.GET("/member/add/:id", groupController.AddMember)
		groupGroup.POST("/member/role/:id", groupController.UpdateMemberRole)
		groupGroup.GET("/member/remove/:id", groupController.RemoveMember)
		groupGroup.GET("/member/list/:id", groupController.ListMembers)
	}
}

// 注册项目相关路由
func registerProjectRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, jwtMiddleware *middleware.JWTAuthMiddleware, casbinService service.CasbinService) {
	// 初始化项目仓库和服务
	projectRepo := repository.NewProjectRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	userRepo := repository.NewUserRepository(db)

	projectService := service.NewProjectService(
		projectRepo,
		groupRepo,
		userRepo,
		casbinService,
	)
	projectController := controller.NewProjectController(projectService)

	// 项目相关路由
	projectGroup := apiGroup.Group("/project")
	projectGroup.Use(jwtMiddleware.AuthMiddleware())
	{
		// 项目管理
		projectGroup.POST("/create", projectController.CreateProject)
		projectGroup.POST("/update", projectController.UpdateProject)
		projectGroup.GET("/detail/:id", projectController.GetProjectByID)
		projectGroup.GET("/delete/:id", projectController.DeleteProject)
		projectGroup.GET("/list", projectController.ListProjects)
		projectGroup.GET("/user", projectController.GetUserProjects)

		// 项目成员管理
		projectGroup.POST("/permission/set", projectController.SetPermission)
		projectGroup.POST("/permission/remove", projectController.RemovePermission)
		projectGroup.GET("/users/:id", projectController.ListProjectUsers)
	}
}

// 注册文件相关路由 (空壳)
func registerFileRoutes(apiGroup *gin.RouterGroup, jwtMiddleware *middleware.JWTAuthMiddleware) {
	fileGroup := apiGroup.Group("/file")
	fileGroup.Use(jwtMiddleware.AuthMiddleware())
	{
		// 文件相关路由 (空壳)
		fileGroup.POST("/upload", func(c *gin.Context) {})
		fileGroup.GET("/download/:id", func(c *gin.Context) {})
		fileGroup.GET("/delete/:id", func(c *gin.Context) {})
		fileGroup.GET("/list", func(c *gin.Context) {})
	}
}
