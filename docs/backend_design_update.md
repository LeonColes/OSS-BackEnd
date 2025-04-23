# OSS-Backend 后端设计更新文档 - 权限部分

## 权限系统设计更新

原有的权限系统设计进行了简化和优化，主要更新包括：

1. 简化角色体系
2. 支持用户自主创建群组
3. 采用标签化权限管理
4. 统一权限验证中间件

### 1. 简化角色设计

系统角色从原来的四级角色（系统管理员、群组管理员、项目管理员、普通成员）简化为三级角色：

| 角色名称 | 角色编码 | 职责概述 |
|---------|----------|---------|
| 系统管理员 | SUPER_ADMIN | 拥有系统全局权限，管理所有群组和资源 |
| 群组管理员 | GROUP_ADMIN | 管理群组资源、项目和成员权限 |
| 普通成员 | MEMBER | 根据分配的权限参与群组协作 |

这种简化设计降低了系统复杂度，并使权限管理更加直观。

### 2. 群组创建与管理流程

#### 2.1 群组创建

群组创建改为允许任何已注册用户自主创建：

1. 用户登录系统后可直接创建群组
2. 创建者自动成为该群组的群组管理员
3. 系统可通过配额限制单用户创建群组数量

#### 2.2 成员管理

群组管理员可以：
1. 邀请用户加入群组（直接邀请或生成邀请码）
2. 移除群组成员
3. 通过权限标签直接管理成员权限，无需提升角色

### 3. 权限标签管理

采用权限标签替代角色提升，实现更细粒度的权限控制：

| 权限标签 | 说明 | 默认分配 |
|---------|------|---------|
| read | 读取权限，可查看文件 | 所有成员 |
| upload | 上传权限，可上传新文件 | 所有成员 |
| download | 下载权限，可下载文件 | 所有成员 |
| delete | 删除权限，可删除文件 | 特定成员 |
| share | 分享权限，可分享文件给外部用户 | 特定成员 |
| admin | 管理权限，可管理项目设置 | 仅管理员 |

通过组合这些标签，可以实现丰富多样的权限设置。

### 4. 统一权限中间件

新设计引入了统一的权限验证中间件，替代原来的三个中间件：

```go
// 统一权限验证中间件
func (m *UnifiedCasbinMiddleware) AuthCheck(level string, resourceType string) gin.HandlerFunc {
    // 根据不同权限级别执行相应验证逻辑:
    // - 系统级: 验证用户对系统API的访问权限
    // - 群组级: 验证用户对特定群组资源的操作权限
    // - 项目级: 验证用户对特定项目资源的操作权限
}
```

使用方式简化为：

```go
// 示例 - 系统级权限验证
router.GET("/api/oss/user/list", 
    authzMiddleware.AuthCheck(middleware.LevelSystem, ""))

// 示例 - 群组级权限验证（文件资源）
router.POST("/api/oss/group/:id/file/upload", 
    authzMiddleware.AuthCheck(middleware.LevelGroup, "file"))
```

### 5. 代码结构优化

权限相关代码结构进行了优化：

1. 工具类函数移至统一的`utils`包
2. 中间件按功能分组存放在子目录中
3. 统一错误处理和响应格式

```
internal/
  ├── utils/                 # 工具类函数
  │   └── http_utils.go      # HTTP相关工具函数
  ├── middleware/            # 中间件
  │   ├── auth/              # 认证授权中间件
  │   │   ├── jwt.go         # JWT认证中间件
  │   │   └── casbin.go      # 统一Casbin权限中间件
  │   └── ...
  ├── service/
  │   └── casbin_service.go  # Casbin服务
  └── ...
```

### 6. 权限系统API

权限系统提供以下几类API:

| 分类 | 说明 | 示例 |
|-----|------|------|
| 角色管理 | 分配和撤销用户角色 | `/api/oss/role/assign`, `/api/oss/role/revoke` |
| 权限管理 | 授予和撤销权限 | `/api/oss/permission/grant`, `/api/oss/permission/revoke` |
| 群组管理 | 创建和管理群组 | `/api/oss/group/create`, `/api/oss/group/my-groups` |
| 成员管理 | 管理群组成员 | `/api/oss/group/{id}/invite`, `/api/oss/group/{id}/member/{user_id}` |
| 权限检查 | 验证权限 | `/api/oss/permission/check`, `/api/oss/permission/list` |

详细API文档和使用方法参见 `permission_api_guide.md`。 