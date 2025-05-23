# API文档

## 接口规范

### 基本信息

- 基础路径: `/api/oss`
- 所有接口均采用RESTful设计风格
- 接口版本通过URL路径指定，如`/api/oss/v1/users`

### 请求格式

- 请求格式支持JSON和FormData两种方式
- Content-Type: `application/json` 或 `multipart/form-data`（文件上传）
- 接口鉴权方式：Bearer Token（通过Authorization头传递）
- 文件上传使用FormData格式

### 响应格式

所有接口统一返回以下格式的JSON:

```json
{
  "code": 0,          // 0表示成功，非0表示失败
  "message": "成功",   // 响应消息
  "data": {}          // 响应数据，成功时返回
}
```

### 错误码定义

| 错误码 | 描述 |
|--------|------|
| 0 | 成功 |
| 10001 | 参数验证失败 |
| 10002 | 资源不存在 |
| 10003 | 资源已存在 |
| 20001 | 用户未登录 |
| 20002 | 登录失败 |
| 20003 | Token无效或过期 |
| 30001 | 权限不足 |
| 40001 | 服务器内部错误 |
| 50001 | 文件上传失败 |
| 50002 | 文件下载失败 |
| 50003 | 文件过大 |

### 分页参数

支持分页的接口统一使用以下查询参数:

- `page`: 页码，从1开始
- `size`: 每页条数，默认20，最大100
- `sort`: 排序字段，如`created_at`
- `order`: 排序方向，asc或desc

分页响应格式:

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "list": [],        // 数据列表
    "total": 0,        // 总记录数
    "page": 1,         // 当前页码
    "size": 20,        // 每页条数
    "pages": 1         // 总页数
  }
}
```

## 角色权限列表

### 角色定义

系统中有三种预定义角色：
1. **ADMIN（系统管理员）** - 拥有全局权限
2. **GROUP_ADMIN（群组管理员）** - 拥有群组级别的管理权限
3. **MEMBER（普通成员）** - 拥有基本使用权限

### 权限清单

| API接口路径 | ADMIN | GROUP_ADMIN | MEMBER | 说明 |
|------------|-------|-------------|--------|------|
| **/api/oss/user/register** | ✓ | ✓ | ✓ | 用户注册（公开） |
| **/api/oss/user/login** | ✓ | ✓ | ✓ | 用户登录（公开） |
| **/api/oss/user/info** | ✓ | ✓ | ✓ | 获取个人信息（需登录） |
| **/api/oss/user/update** | ✓ | ✓ | ✓ | 更新个人信息（需登录） |
| **/api/oss/user/password** | ✓ | ✓ | ✓ | 修改密码（需登录） |
| **/api/oss/user/list** | ✓ | ✓ | ✗ | 用户列表（需要GROUP_ADMIN权限） |
| **/api/oss/user/status/:id** | ✓ | ✓ | ✗ | 更新用户状态（需要GROUP_ADMIN权限） |
| **/api/oss/user/roles/:id** | ✓ | ✓ | ✗ | 获取用户角色（需要GROUP_ADMIN权限） |
| **/api/oss/user/roles/:id** (POST) | ✓ | ✓ | ✗ | 分配用户角色（需要GROUP_ADMIN权限） |
| **/api/oss/user/roles/:id/remove** | ✓ | ✓ | ✗ | 移除用户角色（需要GROUP_ADMIN权限） |
| **/api/oss/role/create** | ✓ | ✓ | ✗ | 创建角色（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/update** | ✓ | ✓ | ✗ | 更新角色（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/delete/:id** | ✓ | ✓ | ✗ | 删除角色（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/detail/:id** | ✓ | ✓ | ✗ | 角色详情（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/list** | ✓ | ✓ | ✗ | 角色列表（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/group/create** | ✓ | ✓ | ✓ | 创建群组（需登录） |
| **/api/oss/group/update** | ✓ | ✓ | ✗ | 更新群组（需要GROUP_ADMIN权限） |
| **/api/oss/group/detail/:id** | ✓ | ✓ | ✓ | 群组详情（需登录） |
| **/api/oss/group/list** | ✓ | ✓ | ✓ | 群组列表（需登录） |
| **/api/oss/group/user** | ✓ | ✓ | ✓ | 获取用户所在群组（需登录） |
| **/api/oss/group/join** | ✓ | ✓ | ✓ | 加入群组（需登录） |
| **/api/oss/group/invite** | ✓ | ✓ | ✓ | 生成邀请码（需登录） |
| **/api/oss/group/member/add/:id** | ✓ | ✓ | ✗ | 添加成员（需要GROUP_ADMIN权限） |
| **/api/oss/group/member/role/:id** | ✓ | ✓ | ✗ | 更新成员角色（需要GROUP_ADMIN权限） |
| **/api/oss/group/member/remove/:id** | ✓ | ✓ | ✗ | 移除成员（需要GROUP_ADMIN权限） |
| **/api/oss/group/member/list/:id** | ✓ | ✓ | ✗ | 成员列表（需要GROUP_ADMIN权限） |
| **/api/oss/project/create** | ✓ | ✓ | ✗ | 创建项目（需要GROUP_ADMIN角色） |
| **/api/oss/project/update** | ✓ | ✓ | ✗ | 更新项目（需要项目/群组权限） |
| **/api/oss/project/detail/:id** | ✓ | ✓ | ✓ | 项目详情（需要读取权限） |
| **/api/oss/project/delete/:id** | ✓ | ✓ | ✗ | 删除项目（需要项目/群组权限） |
| **/api/oss/project/list** | ✓ | ✓ | ✓ | 项目列表（需要读取权限） |
| **/api/oss/project/user** | ✓ | ✓ | ✓ | 获取用户项目（需登录） |
| **/api/oss/project/member/add** | ✓ | ✓ | ✗ | 添加项目成员（需要GROUP_ADMIN权限） |
| **/api/oss/project/member/remove** | ✓ | ✓ | ✗ | 移除项目成员（需要GROUP_ADMIN权限） |
| **/api/oss/project/member/list/:id** | ✓ | ✓ | ✗ | 项目成员列表（需要GROUP_ADMIN权限） |
| **/api/oss/file/upload** | ✓ | ✓ | ✓ | 上传文件（需要create文件权限） |
| **/api/oss/file/download/:id** | ✓ | ✓ | ✓ | 下载文件（需要read文件权限） |
| **/api/oss/file/list** | ✓ | ✓ | ✓ | 文件列表（需要read文件权限） |
| **/api/oss/file/delete/:id** | ✓ | ✓ | ✓ | 删除文件（需要delete文件权限） |

## 核心接口说明

### 身份认证

#### 用户注册

```
POST /api/oss/v1/auth/register
```

请求体:
```json
{
  "username": "testuser",
  "password": "Password123!",
  "email": "test@example.com",
  "name": "测试用户"
}
```

#### 用户登录

```
POST /api/oss/v1/auth/login
```

请求体:
```json
{
  "username": "testuser",
  "password": "Password123!"
}
```

响应:
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJ...",
    "expire_at": 1679984523,
    "user": {
      "id": 1,
      "username": "testuser",
      "name": "测试用户",
      "email": "test@example.com",
      "role": "MEMBER"
    }
  }
}
```

### 用户管理

#### 获取用户列表

```
GET /api/oss/v1/users
```

权限要求: `ADMIN`

#### 获取用户详情

```
GET /api/oss/v1/users/{id}
```

权限要求: `ADMIN` 或 当前用户

### 群组管理

#### 创建群组

```
POST /api/oss/group/create
```

请求体:
```json
{
  "name": "测试群组",
  "description": "这是一个测试群组",
  "group_key": "test_group"  // 仅允许字母和数字，作为MinIO桶名
}
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": null
}
```

权限要求: 任何已登录用户

#### 更新群组信息

```
POST /api/oss/group/update
```

请求体:
```json
{
  "id": 1,
  "name": "更新后的群组名称",
  "description": "更新后的群组描述",
  "status": 1  // 可选，1-正常, 2-禁用, 3-锁定
}
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": null
}
```

权限要求: 群组管理员 (需要有群组管理员权限)

#### 获取群组详情

```
GET /api/oss/group/detail/{id}
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 1,
    "name": "测试群组",
    "description": "这是一个测试群组",
    "group_key": "test_group",
    "invite_code": "ABC123",  // 仅群组管理员可见
    "storage_quota": 0,  // 存储配额，0表示无限制
    "storage_used": 1024,  // 已使用的存储空间（字节）
    "member_count": 10,  // 成员数量
    "project_count": 5,  // 项目数量
    "status": 1,  // 1-正常, 2-禁用, 3-锁定
    "creator_id": 1,
    "creator_name": "创建者",
    "created_at": "2023-06-01T12:00:00Z",
    "user_role": "admin"  // 当前用户在群组中的角色
  }
}
```

权限要求: 群组成员

#### 获取群组列表

```
GET /api/oss/group/list?name=&status=1&page=1&size=10
```

参数:
- `name`: 群组名称（模糊查询，可选）
- `status`: 状态（1-正常, 2-禁用, 3-锁定，可选）
- `page`: 页码，默认1
- `size`: 每页数量，默认10

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 100,
    "items": [
      {
        "id": 1,
        "name": "测试群组1",
        "description": "这是测试群组1",
        "group_key": "test_group_1",
        "storage_quota": 0,
        "storage_used": 1024,
        "member_count": 10,
        "project_count": 5,
        "status": 1,
        "creator_id": 1,
        "creator_name": "创建者",
        "created_at": "2023-06-01T12:00:00Z",
        "user_role": "admin"
      },
      // ... 更多群组
    ]
  }
}
```

权限要求: 任何已登录用户

#### 获取用户所在的群组

```
GET /api/oss/group/user
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "id": 1,
      "name": "测试群组1",
      "description": "这是测试群组1",
      "group_key": "test_group_1",
      "storage_quota": 0,
      "storage_used": 1024,
      "member_count": 10,
      "project_count": 5,
      "status": 1,
      "creator_id": 1,
      "creator_name": "创建者",
      "created_at": "2023-06-01T12:00:00Z",
      "user_role": "admin"
    },
    // ... 更多群组
  ]
}
```

权限要求: 任何已登录用户

#### 加入群组

```
POST /api/oss/group/join
```

请求体:
```json
{
  "invite_code": "ABC123"
}
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": null
}
```

权限要求: 任何已登录用户

#### 生成群组邀请码

```
POST /api/oss/group/invite
```

请求体:
```json
{
  "group_id": 1,
  "expire_days": 30  // 过期天数，0表示永不过期
}
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "group_id": 1,
    "group_name": "测试群组",
    "invite_code": "ABC123",
    "expire_at": "2023-07-01T12:00:00Z"
  }
}
```

权限要求: 群组管理员

### 群组成员管理

#### 添加群组成员

```
GET /api/oss/group/member/add/{id}?user_id=2&role=member
```

参数:
- `id`: 群组ID
- `user_id`: 用户ID
- `role`: 角色，可选值：admin(管理员)、member(普通成员)

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": null
}
```

权限要求: 群组管理员

#### 更新成员角色

```
POST /api/oss/group/member/role/{id}
```

参数:
- `id`: 群组ID

请求体:
```json
{
  "user_id": 2,
  "role": "admin"  // 可选值：admin, member
}
```

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": null
}
```

权限要求: 群组管理员

#### 移除群组成员

```
GET /api/oss/group/member/remove/{id}?user_id=2
```

参数:
- `id`: 群组ID
- `user_id`: 要移除的用户ID

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": null
}
```

权限要求: 群组管理员

#### 获取群组成员列表

```
GET /api/oss/group/member/list/{id}?page=1&size=10
```

参数:
- `id`: 群组ID
- `page`: 页码，默认1
- `size`: 每页数量，默认10

响应:
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 10,
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "user_name": "用户1",
        "email": "user1@example.com",
        "avatar": "https://avatar.url/user1.jpg",
        "role": "admin",
        "joined_at": "2023-06-01T12:00:00Z",
        "last_active_at": "2023-06-02T15:30:00Z"
      },
      // ... 更多成员
    ]
  }
}
```

权限要求: 群组管理员

### 文件管理

#### 上传文件

```
POST /api/oss/v1/groups/{group_id}/projects/{project_id}/files
```

Content-Type: `multipart/form-data`

表单字段:
- `file`: 文件数据
- `path`: 文件路径 (可选)
- `description`: 文件描述 (可选)

权限要求: 对项目有写权限的成员

#### 下载文件

```
GET /api/oss/v1/groups/{group_id}/projects/{project_id}/files/{file_id}/download
```

权限要求: 对项目有读权限的成员

#### 删除文件

```
DELETE /api/oss/v1/groups/{group_id}/projects/{project_id}/files/{file_id}
```

权限要求: 对项目有写权限的成员或文件所有者

## Swagger使用指南

### 访问Swagger文档

开发环境下，可通过浏览器访问Swagger UI:

```
http://localhost:8080/swagger/index.html
```

### 接口测试步骤

1. 在Swagger UI首页，点击右上角的"Authorize"按钮

2. 输入获取到的JWT令牌:
   ```
   Bearer eyJhbGciOiJIUzI1NiIs...
   ```
   
3. 点击"Authorize"按钮完成认证

4. 选择需要测试的接口，点击"Try it out"

5. 填写请求参数，点击"Execute"发送请求

6. 查看响应结果

### API文档更新

Swagger文档通过代码注释自动生成，要更新文档:

1. 在对应的控制器代码中添加Swagger注释

   ```go
   // CreateUser godoc
   // @Summary 创建用户
   // @Description 创建一个新用户
   // @Tags 用户管理
   // @Accept json
   // @Produce json
   // @Param user body UserCreateRequest true "用户信息"
   // @Success 200 {object} Response{data=User} "成功"
   // @Failure 400 {object} Response "请求参数错误"
   // @Failure 403 {object} Response "权限不足"
   // @Router /users [post]
   // @Security BearerAuth
   func (c *UserController) CreateUser(ctx *gin.Context) {
       // 实现代码
   }
   ```

2. 重新生成Swagger文档:

   ```bash
   swag init
   ```

3. 重启应用后访问Swagger UI查看更新后的文档

### 常见问题

#### 接口权限不足

确保已经正确授权:
1. 通过登录接口获取最新的token
2. 在Swagger UI中重新授权
3. 确认当前用户有对应接口的权限

#### 找不到特定接口

可能的原因:
1. 接口尚未添加Swagger注释
2. 需要重新生成Swagger文档
3. 接口路径或方法不正确 