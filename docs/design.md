# OSS-Backend 系统设计文档

## 1. 系统概述

OSS-Backend 是一个企业级对象存储系统的后端服务，提供用户管理、角色管理、项目管理、文件管理等功能。系统采用Go语言开发，基于RESTful API设计风格，使用Gin框架实现HTTP服务，使用GORM作为ORM框架操作数据库。

## 2. 核心功能模块

### 2.1 用户模块

用户模块负责管理系统用户，包括用户注册、登录、信息维护等功能。

#### 核心流程：

1. **用户注册**：
   - 用户提供邮箱、密码和姓名等基本信息
   - 系统验证邮箱是否已被注册
   - 密码加密存储
   - 创建用户账号

2. **用户登录**：
   - 用户提供邮箱和密码
   - 系统验证用户信息
   - 生成JWT令牌返回给客户端
   - 记录登录信息（IP、时间）

3. **用户信息管理**：
   - 更新个人资料
   - 修改密码
   - 查看个人信息

4. **用户角色管理**：
   - 为用户分配角色
   - 移除用户角色
   - 查看用户拥有的角色

### 2.2 角色模块

角色模块负责管理系统角色，支持角色的创建、更新、删除和查询。

#### 预定义角色：

- GROUP_ADMIN：群组管理员
- PROJECT_ADMIN：项目管理员
- MEMBER：普通成员
- UPLOADER：上传者
- READER：只读用户

### 2.3 项目模块（待实现）

项目模块负责管理文件组织结构，一个项目可以包含多个文件。

#### 核心流程：

1. **项目创建**：
   - 用户创建新项目，指定项目名称、描述等信息
   - 项目归属于特定群组

2. **项目管理**：
   - 更新项目信息
   - 删除项目
   - 查看项目详情
   - 列出项目列表

### 2.4 文件模块（待实现）

文件模块负责文件的上传、下载、删除和查询等操作。

#### 核心流程：

1. **文件上传**：
   - 用户上传文件到指定项目
   - 系统存储文件元数据
   - 文件内容存储到对象存储系统

2. **文件下载**：
   - 用户请求下载特定文件
   - 系统验证权限
   - 提供文件下载链接

3. **文件管理**：
   - 删除文件
   - 查看文件列表
   - 更新文件信息

## 3. 数据模型

### 3.1 用户实体 (User)

用户实体描述系统用户信息：

```go
type User struct {
    ID           uint       // 用户ID
    Email        string     // 用户邮箱，登录凭证
    Name         string     // 用户姓名
    PasswordHash string     // 密码哈希值
    Avatar       string     // 用户头像URL
    Status       int        // 状态（1-正常，2-禁用，3-锁定）
    LastLoginAt  *time.Time // 最后登录时间
    LastLoginIP  string     // 最后登录IP
    CreatedAt    time.Time  // 创建时间
    UpdatedAt    time.Time  // 更新时间
    
    // 用户角色关联（多对多）
    Roles []Role  // 用户角色
}
```

### 3.2 角色实体 (Role)

角色实体描述系统角色信息：

```go
type Role struct {
    ID          uint      // 角色ID
    Name        string    // 角色名称
    Description string    // 角色描述
    Code        string    // 角色编码，用于权限控制
    Status      int       // 状态：1-启用，0-禁用
    IsSystem    bool      // 是否为系统角色，系统角色不可删除
    CreatedBy   uint      // 创建者ID
    UpdatedBy   uint      // 更新者ID
    CreatedAt   time.Time // 创建时间
    UpdatedAt   time.Time // 更新时间
}
```

### 3.3 用户角色关联 (UserRole)

用户与角色的多对多关联：

```go
type UserRole struct {
    UserID    uint      // 用户ID
    RoleID    uint      // 角色ID
    CreatedAt time.Time // 创建时间
}
```

### 3.4 项目实体 (Project)

项目实体描述文件组织结构：

```go
type Project struct {
    ID          uint64      // 项目ID
    GroupID     uint64      // 群组ID
    Name        string      // 项目名称
    Description string      // 项目描述
    PathPrefix  string      // 路径前缀
    CreatorID   uint64      // 创建者ID
    CreatedAt   time.Time   // 创建时间
    UpdatedAt   time.Time   // 更新时间
    Status      int         // 状态：1-正常, 2-归档, 3-删除
    DeletedAt   gorm.DeletedAt // 删除时间
}
```

### 3.5 权限实体 (Permission)

权限实体描述用户对项目的权限：

```go
type Permission struct {
    ID        uint64     // 权限ID
    UserID    uint64     // 用户ID
    ProjectID uint64     // 项目ID
    Role      string     // 角色：admin, editor, viewer
    CreatedAt time.Time  // 创建时间
    UpdatedAt time.Time  // 更新时间
    ExpireAt  *time.Time // 过期时间
    GrantedBy uint64     // 授权者ID
}
```

## 4. API接口规范

### 4.1 响应格式

所有API接口返回统一的JSON格式：

```json
{
    "code": 200,        // 状态码：200成功，非200表示失败
    "message": "success", // 消息说明
    "data": {}          // 返回的数据，可选
}
```

### 4.2 状态码说明

- 200: 成功
- 400: 请求参数错误
- 401: 未授权
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器内部错误

### 4.3 认证方式

系统使用JWT (JSON Web Token) 进行身份认证：

1. 用户登录后获取token
2. 后续请求在Header中携带token：`Authorization: Bearer {token}`
3. 服务器验证token有效性
4. token过期后，使用refresh_token获取新token

## 5. 系统流程

### 5.1 注册登录流程

1. **注册流程**：
   - 用户访问注册接口 `/api/oss/user/register`
   - 提供邮箱、密码和姓名
   - 系统验证邮箱是否已存在
   - 密码加密并创建用户
   - 返回成功消息

2. **登录流程**：
   - 用户访问登录接口 `/api/oss/user/login`
   - 提供邮箱和密码
   - 系统验证用户信息
   - 生成JWT令牌
   - 返回令牌和用户信息

3. **身份验证流程**：
   - 用户请求需要认证的接口
   - 在请求头中携带令牌
   - 中间件验证令牌有效性
   - 解析令牌获取用户信息
   - 将用户信息写入请求上下文

### 5.2 权限控制流程

1. **角色分配**：
   - 管理员访问角色分配接口 `/api/oss/user/roles/{id}`
   - 为指定用户分配角色
   - 系统建立用户和角色的关联

2. **权限检查**：
   - 用户请求需要特定权限的接口
   - 中间件检查用户拥有的角色
   - 根据角色权限决定是否允许访问
   - 权限不足则返回403错误

### 5.3 文件上传流程（待实现）

1. **文件上传**：
   - 用户选择文件并指定上传目标项目
   - 前端调用上传接口 `/api/oss/file/upload`
   - 后端验证用户权限
   - 生成文件元数据并保存到数据库
   - 文件内容保存到对象存储系统
   - 返回文件信息

2. **文件下载**：
   - 用户请求下载文件 `/api/oss/file/download/{id}`
   - 系统验证用户是否有权限访问该文件
   - 从对象存储系统获取文件内容
   - 返回文件下载流

## 6. 安全设计

1. **密码安全**：
   - 使用bcrypt算法加密存储密码
   - 不存储明文密码
   - 密码传输过程中使用HTTPS加密

2. **认证安全**：
   - 使用JWT进行身份认证
   - 令牌设置合理过期时间
   - 使用refresh_token机制更新令牌

3. **权限控制**：
   - 基于角色的访问控制
   - 接口级别权限检查
   - 数据级别权限校验

## 7. 部署方案

### 7.1 开发环境

- 使用SQLite内存数据库方便开发和测试
- 使用Make命令管理构建和运行
- Swagger自动生成API文档

### 7.2 生产环境（规划）

- 使用MySQL作为持久化数据库
- 使用Redis缓存会话和热点数据
- 使用MinIO作为对象存储系统
- 使用Docker容器化部署
- 使用Docker Compose编排服务 