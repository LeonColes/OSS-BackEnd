package routes

import (
	"context"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "oss-backend/docs/swagger" // 导入Swagger文档
	"oss-backend/internal/controller"
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

	// 设置 API 前缀
	apiGroup := r.Group("/api/oss")

	// 注册角色相关路由
	registerRoleRoutes(apiGroup, gormDB)

	// 注册用户相关路由
	registerUserRoutes(apiGroup, gormDB)

	// 注册项目相关路由 (空壳)
	registerProjectRoutes(apiGroup)

	// 注册文件相关路由 (空壳)
	registerFileRoutes(apiGroup)
}

// 注册角色相关路由
func registerRoleRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	// 创建依赖
	roleRepo := repository.NewRoleRepository(db)
	roleService := service.NewRoleService(roleRepo)
	roleController := controller.NewRoleController(roleService)

	// 初始化系统角色
	err := roleService.InitSystemRoles(context.Background())
	if err != nil {
		panic("初始化系统角色失败: " + err.Error())
	}

	// 角色相关路由
	roleGroup := apiGroup.Group("/role")
	{
		roleGroup.POST("/create", roleController.CreateRole)
		roleGroup.POST("/update", roleController.UpdateRole)
		roleGroup.GET("/delete/:id", roleController.DeleteRole)
		roleGroup.GET("/detail/:id", roleController.GetRoleByID)
		roleGroup.GET("/list", roleController.ListRoles)
	}
}

// 注册用户相关路由
func registerUserRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
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

		// 需要认证的路由
		userGroup.GET("/info", userController.GetUserInfo)
		userGroup.POST("/update", userController.UpdateUserInfo)
		userGroup.POST("/password", userController.UpdatePassword)
		userGroup.GET("/list", userController.ListUsers)
		userGroup.GET("/status/:id", userController.UpdateUserStatus)

		// 用户角色管理
		userGroup.GET("/roles/:id", userController.GetUserRoles)
		userGroup.POST("/roles/:id", userController.AssignRoles)
		userGroup.POST("/roles/:id/remove", userController.RemoveRoles)
	}
}

// 注册项目相关路由 (空壳)
func registerProjectRoutes(apiGroup *gin.RouterGroup) {
	projectGroup := apiGroup.Group("/project")
	{
		// 项目相关路由 (空壳)
		projectGroup.POST("/create", func(c *gin.Context) {})
		projectGroup.POST("/update", func(c *gin.Context) {})
		projectGroup.GET("/delete/:id", func(c *gin.Context) {})
		projectGroup.GET("/detail/:id", func(c *gin.Context) {})
		projectGroup.GET("/list", func(c *gin.Context) {})
	}
}

// 注册文件相关路由 (空壳)
func registerFileRoutes(apiGroup *gin.RouterGroup) {
	fileGroup := apiGroup.Group("/file")
	{
		// 文件相关路由 (空壳)
		fileGroup.POST("/upload", func(c *gin.Context) {})
		fileGroup.GET("/download/:id", func(c *gin.Context) {})
		fileGroup.GET("/delete/:id", func(c *gin.Context) {})
		fileGroup.GET("/list", func(c *gin.Context) {})
	}
}
