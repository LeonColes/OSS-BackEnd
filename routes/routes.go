package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "oss-backend/docs" // 导入 Swagger 文档
	"oss-backend/internal/controller"
	"oss-backend/internal/repository"
	"oss-backend/internal/service"
)

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine, db interface{}) {
	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 设置 API 前缀
	apiGroup := r.Group("/api/oss")

	// 注册角色相关路由
	registerRoleRoutes(apiGroup)

	// 注册用户相关路由 (空壳)
	registerUserRoutes(apiGroup)

	// 注册项目相关路由 (空壳)
	registerProjectRoutes(apiGroup)

	// 注册文件相关路由 (空壳)
	registerFileRoutes(apiGroup)
}

// 注册角色相关路由
func registerRoleRoutes(apiGroup *gin.RouterGroup) {
	// 创建依赖
	// 注意: 这里临时使用 nil 替代数据库连接，实际项目中应该传入真实的 DB 实例
	roleRepo := repository.NewRoleRepository(nil)
	roleService := service.NewRoleService(roleRepo)
	roleController := controller.NewRoleController(roleService)

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

// 注册用户相关路由 (空壳)
func registerUserRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/user")
	{
		// 用户相关路由 (空壳)
		userGroup.POST("/register", func(c *gin.Context) {})
		userGroup.POST("/login", func(c *gin.Context) {})
		userGroup.GET("/info", func(c *gin.Context) {})
		userGroup.POST("/update", func(c *gin.Context) {})
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
