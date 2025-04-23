# OSS-Backend 角色设计文档

## 1. 角色设计概述

为满足企业级对象存储系统的权限管理需求，系统采用基于角色的访问控制(RBAC)模型，根据业务需求设计了三个核心角色：

| 角色名称 | 角色编码 | 职责概述 |
|---------|----------|---------|
| 群组管理员 | GROUP_ADMIN | 负责管理和规划项目，控制项目数量和资源分配 |
| 项目管理员 | PROJECT_ADMIN | 负责管理项目成员，控制项目人数和成员权限 |
| 普通成员 | MEMBER | 具有基本的文件操作权限，可以参与项目协作 |

## 2. 角色权限矩阵

| 功能/操作 | 群组管理员 | 项目管理员 | 普通成员 |
|----------|-----------|-----------|---------|
| **用户管理** |  |  |  |
| 查看个人信息 | ✓ | ✓ | ✓ |
| 修改个人信息 | ✓ | ✓ | ✓ |
| 查看用户列表 | ✓ | ✓ | ✗ |
| 查看群组列表 | ✓ | ✗ | ✗ |
| **项目管理** |  |  |  |
| 创建项目 | ✓ | ✗ | ✗ |
| 编辑项目信息 | ✓ | ✓ | ✗ |
| 查看所有项目 | ✓ | ✗ | ✗ |
| 查看参与项目 | ✓ | ✓ | ✓ |
| 删除项目(逻辑) | ✓ | ✗ | ✗ |
| **成员管理** |  |  |  |
| 添加项目成员 | ✓ | ✓ | ✗ |
| 移除项目成员 | ✓ | ✓ | ✗ |
| 分配角色 | ✓ | ✓ | ✗ |
| **文件管理** |  |  |  |
| 上传文件 | ✓ | ✓ | ✓ |
| 下载文件 | ✓ | ✓ | ✓ |
| 删除文件(逻辑) | ✓ | ✓ | ✓ |
| 查看文件列表 | ✓ | ✓ | ✓ |

## 3. 角色实现机制

### 3.1 数据模型

```go
// Role 角色实体
type Role struct {
    ID          uint      // 角色ID
    Name        string    // 角色名称
    Description string    // 角色描述
    Code        string    // 角色编码
    Status      int       // 状态：1-启用，0-禁用
    IsSystem    bool      // 系统角色标识，系统角色不可删除
    CreatedBy   uint      // 创建者ID
    UpdatedBy   uint      // 更新者ID
    CreatedAt   time.Time // 创建时间
    UpdatedAt   time.Time // 更新时间
}

// 预定义系统角色
const (
    RoleGroupAdmin   = "GROUP_ADMIN"   // 群组管理员
    RoleProjectAdmin = "PROJECT_ADMIN" // 项目管理员
    RoleMember       = "MEMBER"        // 普通成员
)
```

### 3.2 角色初始化

系统在初始启动时，会通过`InitSystemRoles`方法自动创建这三个基本角色：

```go
// InitSystemRoles 初始化系统角色
func (r *roleRepository) InitSystemRoles(ctx context.Context) error {
    // 预定义的系统角色
    systemRoles := []*entity.Role{
        {
            Name:        "群组管理员",
            Description: "负责管理和规划项目，控制项目数量和资源分配",
            Code:        entity.RoleGroupAdmin,
            Status:      1,
            IsSystem:    true,
        },
        {
            Name:        "项目管理员",
            Description: "负责管理项目成员，控制项目人数和成员权限",
            Code:        entity.RoleProjectAdmin,
            Status:      1,
            IsSystem:    true,
        },
        {
            Name:        "普通成员",
            Description: "具有基本的文件操作权限，可以参与项目协作",
            Code:        entity.RoleMember,
            Status:      1,
            IsSystem:    true,
        },
    }
    
    // 事务处理，检查是否已存在，不存在则创建
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 代码实现...
    })
}
```

## 4. 角色分配流程

### 4.1 新用户角色分配

1. 用户注册过程：
   - 用户提交注册信息
   - 系统验证信息并创建用户账号
   - 系统自动为新用户分配"普通成员(MEMBER)"角色

2. 代码实现：
```go
// Register 用户注册
func (s *userService) Register(ctx context.Context, req *dto.UserRegisterRequest) error {
    // 创建用户...
    
    // 为新用户分配普通成员角色
    memberRole, err := s.roleRepo.GetByCode(ctx, entity.RoleMember)
    if err == nil && memberRole != nil {
        _ = s.userRepo.AssignRoles(ctx, uint64(user.ID), []uint{memberRole.ID})
    }
    
    return nil
}
```

### 4.2 角色升级流程

1. 普通成员升级为项目管理员：
   ```
   群组管理员 -> 查看项目成员 -> 选择成员 -> 分配"项目管理员"角色 -> 确认
   ```

2. 项目管理员降级为普通成员：
   ```
   群组管理员 -> 查看项目成员 -> 选择成员 -> 移除"项目管理员"角色 -> 确认
   ```

3. 任命群组管理员（需要超级管理员权限）：
   ```
   系统管理员 -> 用户管理 -> 选择用户 -> 分配"群组管理员"角色 -> 确认
   ```

## 5. 权限验证流程

### 5.1 API接口权限验证

1. 请求到达系统：
   ```
   客户端请求 -> JWT认证中间件 -> 角色权限中间件 -> 控制器处理
   ```

2. 认证流程图：
   ```
   +-------------+    +----------------+    +----------------+    +---------------+
   | 客户端请求  | -> | JWT认证中间件  | -> | 角色权限中间件 | -> | 业务处理逻辑  |
   +-------------+    +----------------+    +----------------+    +---------------+
          |                   |                     |                    |
          |                   v                     |                    |
          |           验证Token有效性               |                    |
          |                   |                     v                    |
          |                   |            验证用户是否拥有所需角色       |
          v                   v                     v                    v
   +-------------+    +----------------+    +----------------+    +---------------+
   | 拒绝未授权  |    | 拒绝Token无效  |    | 拒绝权限不足   |    | 返回处理结果  |
   +-------------+    +----------------+    +----------------+    +---------------+
   ```

### 5.2 具体实现

1. JWT认证中间件：
```go
// AuthMiddleware 认证中间件
func (m *JWTAuthMiddleware) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 验证token...
        
        // 解析成功后设置用户ID到上下文
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

2. 角色权限中间件：
```go
// RequireRole 要求特定角色
func (m *RoleAuthMiddleware) RequireRole(roleName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从上下文获取用户ID
        userID := c.GetUint("userID")
        
        // 查询用户角色
        roles, _ := m.userRepo.GetUserRoles(c, uint64(userID))
        
        // 检查是否拥有特定角色
        hasRole := false
        for _, role := range roles {
            if role.Code == roleName {
                hasRole = true
                break
            }
        }
        
        if !hasRole {
            c.JSON(http.StatusForbidden, common.ErrorResponse("权限不足"))
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 6. 角色应用场景

### 6.1 群组管理员

群组管理员负责:
- 创建和规划项目
- 设置项目配额和资源限制
- 任命项目管理员
- 监控项目整体运行状况

典型操作流程:
```
1. 创建新项目
   群组管理员 -> 项目管理 -> 创建项目 -> 填写项目信息 -> 确认

2. 任命项目管理员
   群组管理员 -> 选择项目 -> 成员管理 -> 选择成员 -> 分配项目管理员角色 -> 确认
```

### 6.2 项目管理员

项目管理员负责:
- 管理项目成员
- 控制项目人数
- 分配成员权限
- 管理项目文件

典型操作流程:
```
1. 添加项目成员
   项目管理员 -> 成员管理 -> 添加成员 -> 选择用户 -> 确认

2. 管理文件权限
   项目管理员 -> 文件管理 -> 选择文件 -> 设置权限 -> 确认
```

### 6.3 普通成员

普通成员可以:
- 查看参与的项目
- 上传、下载和管理文件
- 参与项目协作

典型操作流程:
```
1. 上传文件
   普通成员 -> 选择项目 -> 文件管理 -> 上传文件 -> 选择文件 -> 确认

2. 下载文件
   普通成员 -> 选择项目 -> 文件管理 -> 选择文件 -> 下载
```

## 7. 逻辑删除说明

系统采用逻辑删除机制，确保数据可恢复性:

- 项目逻辑删除: 项目状态变更为"已删除"，不会从数据库中物理删除
- 文件逻辑删除: 文件标记为"已删除"，不会从存储中物理删除
- 用户逻辑删除: 用户状态变更为"已禁用"，账号信息保留

实现方式:
```go
// 使用gorm的软删除机制
type Project struct {
    // ...其他字段
    DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除标记
}

// 或使用状态字段标记
type File struct {
    // ...其他字段
    Status int // 1-正常, 2-已删除
}
``` 