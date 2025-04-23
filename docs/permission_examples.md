# 权限中间件使用示例

本文档展示如何在项目中使用统一的权限中间件进行权限控制。

## 1. 初始化中间件

在项目启动阶段，需要初始化权限中间件并注册到路由中：

```go
// 在main.go或routes/router.go中

// 初始化服务
db := initDB()
casbinService, err := service.NewCasbinService(db, userRepo, roleRepo, groupRepo)
if err != nil {
    log.Fatalf("初始化Casbin服务失败: %v", err)
}

// 初始化中间件
unifiedCasbinMiddleware := middleware.NewUnifiedCasbinMiddleware(casbinService)
jwtMiddleware := middleware.NewJWTAuthMiddleware(jwtSecret)

// 初始化路由
r := gin.Default()
initRoutes(r, unifiedCasbinMiddleware, jwtMiddleware)
```

## 2. 注册API路由

使用统一中间件为不同API添加权限校验：

```go
func initRoutes(r *gin.Engine, authz *middleware.UnifiedCasbinMiddleware, jwt *middleware.JWTAuthMiddleware) {
    // 基础API组
    api := r.Group("/api/oss")
    
    // 公共API，无需权限
    api.POST("/user/register", userController.Register)
    api.POST("/user/login", userController.Login)
    
    // 需要认证的API
    authRequired := api.Group("/")
    authRequired.Use(jwt.AuthMiddleware())
    
    // 用户相关 API - 系统级权限校验
    userAPI := authRequired.Group("/user")
    userAPI.Use(authz.AuthCheck(middleware.LevelSystem, ""))
    {
        userAPI.GET("/info", userController.GetUserInfo)
        userAPI.PUT("/password", userController.ChangePassword)
        userAPI.POST("/logout", userController.Logout)
    }
    
    // 群组相关 API - 群组级权限校验
    groupAPI := authRequired.Group("/group")
    {
        // 系统级操作 - 创建群组
        groupAPI.POST("/create", authz.AuthCheck(middleware.LevelSystem, ""), groupController.CreateGroup)
        
        // 群组内操作 - 需要群组级权限校验
        groupAPI.GET("/:id", authz.AuthCheck(middleware.LevelGroup, "group"), groupController.GetGroupInfo)
        groupAPI.PUT("/:id", authz.AuthCheck(middleware.LevelGroup, "group"), groupController.UpdateGroup)
        
        // 群组文件操作
        groupAPI.GET("/:id/files", authz.AuthCheck(middleware.LevelGroup, "file"), fileController.ListGroupFiles)
        groupAPI.POST("/:id/upload", authz.AuthCheck(middleware.LevelGroup, "file"), fileController.UploadFile)
    }
    
    // 项目相关 API - 项目级权限校验
    projectAPI := authRequired.Group("/project")
    {
        // 创建项目 - 系统级权限
        projectAPI.POST("/create", authz.AuthCheck(middleware.LevelSystem, ""), projectController.CreateProject)
        
        // 项目内操作 - 需要项目级权限校验
        projectAPI.GET("/:id", authz.AuthCheck(middleware.LevelProject, "project"), projectController.GetProjectInfo)
        projectAPI.PUT("/:id", authz.AuthCheck(middleware.LevelProject, "project"), projectController.UpdateProject)
        
        // 项目成员管理
        projectAPI.POST("/:id/member", authz.AuthCheck(middleware.LevelProject, "member"), projectController.AddMember)
        projectAPI.DELETE("/:id/member/:uid", authz.AuthCheck(middleware.LevelProject, "member"), projectController.RemoveMember)
        
        // 项目文件操作
        projectAPI.GET("/:id/files", authz.AuthCheck(middleware.LevelProject, "file"), fileController.ListProjectFiles)
        projectAPI.POST("/:id/upload", authz.AuthCheck(middleware.LevelProject, "file"), fileController.UploadProjectFile)
    }
    
    // 权限管理 API - 仅系统管理员可访问
    permissionAPI := authRequired.Group("/permission")
    permissionAPI.Use(authz.AuthCheck(middleware.LevelSystem, ""))
    {
        permissionAPI.POST("/grant", permissionController.GrantPermission)
        permissionAPI.POST("/revoke", permissionController.RevokePermission)
        permissionAPI.GET("/check", permissionController.CheckPermission)
    }
}
```

## 3. 使用场景举例

### 3.1 系统级权限控制

系统级权限用于控制用户对系统级API的访问，例如创建群组、管理用户等。权限验证基于用户角色和API路径。

```go
// 使用示例:
router.POST("/api/oss/group/create", 
    jwtMiddleware.AuthMiddleware(), 
    authzMiddleware.AuthCheck(middleware.LevelSystem, ""),
    groupController.CreateGroup)
```

验证流程:
1. JWT中间件验证用户身份
2. 权限中间件检查用户是否拥有对应角色(如SUPER_ADMIN或GROUP_ADMIN)
3. 如果有权限，执行业务逻辑，否则返回403错误

### 3.2 群组级权限控制

群组级权限用于控制用户对特定群组资源的访问，如群组文件、成员管理等。

```go
// 使用示例:
router.GET("/api/oss/group/:id/files", 
    jwtMiddleware.AuthMiddleware(), 
    authzMiddleware.AuthCheck(middleware.LevelGroup, "file"),
    fileController.ListGroupFiles)
```

验证流程:
1. JWT中间件验证用户身份
2. 权限中间件从路径参数提取群组ID
3. 根据用户ID、群组ID和资源类型(file)检查权限
4. 如果有权限，执行业务逻辑，否则返回403错误

### 3.3 项目级权限控制

项目级权限用于控制用户对特定项目资源的访问，如项目文件、成员管理等。

```go
// 使用示例:
router.POST("/api/oss/project/:id/member", 
    jwtMiddleware.AuthMiddleware(), 
    authzMiddleware.AuthCheck(middleware.LevelProject, "member"),
    projectController.AddMember)
```

验证流程:
1. JWT中间件验证用户身份
2. 权限中间件从路径参数提取项目ID
3. 根据用户ID、项目ID和资源类型(member)检查权限
4. 如果有权限，执行业务逻辑，否则返回403错误

## 4. 自定义资源类型

统一中间件允许指定自定义资源类型，而不是从URL路径提取：

```go
// 指定资源类型为"config"
router.POST("/api/oss/project/:id/settings",
    jwtMiddleware.AuthMiddleware(),
    authzMiddleware.AuthCheck(middleware.LevelProject, "config"),
    projectController.UpdateSettings)
```

这种方式更灵活，可以根据业务需求定义更细粒度的资源类型，而不受API路径结构限制。 