# 系统权限管理说明文档

## 角色定义

系统中有三种预定义角色：
1. **ADMIN（系统管理员）** - 拥有全局权限
2. **GROUP_ADMIN（群组管理员）** - 拥有群组级别的管理权限
3. **MEMBER（普通成员）** - 拥有基本使用权限

## 权限表

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
| **/api/oss/role/create** | ✓ | ✗ | ✗ | 创建角色（需要ADMIN权限） |
| **/api/oss/role/update** | ✓ | ✗ | ✗ | 更新角色（需要ADMIN权限） |
| **/api/oss/role/delete/:id** | ✓ | ✗ | ✗ | 删除角色（需要ADMIN权限） |
| **/api/oss/role/detail/:id** | ✓ | ✗ | ✗ | 角色详情（需要ADMIN权限） |
| **/api/oss/role/list** | ✓ | ✗ | ✗ | 角色列表（需要ADMIN权限） |
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

## 权限关系说明

1. **用户登录流程**：
   - 用户注册时默认分配MEMBER角色
   - 创建群组的用户自动成为该群组的GROUP_ADMIN
   - ADMIN角色通常在系统初始化时创建

2. **资源权限结构**：
   - 系统级权限：全局管理功能（角色管理）
   - 群组级权限：群组内的管理功能（成员管理）
   - 项目级权限：项目内的资源管理（文件操作）

3. **角色权限详细说明**：
   - **ADMIN**：拥有所有系统权限
   - **GROUP_ADMIN**：
     - 可创建/管理项目(projects的create/read/update/delete)
     - 可读取/更新群组(groups的read/update)
     - 可读取用户信息(users的read)
     - 可分配角色(roles的assign)
     - 可管理成员(members的add/remove)
     - 可对文件进行所有操作(files的create/read/update/delete)
   - **MEMBER**：
     - 可读取项目(projects的read)
     - 可对文件进行所有操作(files的create/read/update/delete)

4. **API接口之间的关联**：
   - 用户管理: /api/oss/user/* 处理用户账号、认证和角色管理
   - 角色管理: /api/oss/role/* 处理系统角色定义和权限分配
   - 群组管理: /api/oss/group/* 处理群组创建和成员管理
   - 项目管理: /api/oss/project/* 处理项目及其权限控制
   - 文件管理: /api/oss/file/* 处理文件上传下载等操作

## RBAC权限模型

本系统采用基于角色的访问控制（RBAC）模型，并通过Casbin实现动态权限管理：

- **主体（Subject）**：用户或角色标识，如 "user:1" 或 "ADMIN"
- **对象（Object）**：资源类型，如 "projects", "files" 等
- **动作（Action）**：操作类型，如 "create", "read", "update", "delete"
- **域（Domain）**：权限作用域，如系统域 "system" 或特定群组域 "group:1"

权限策略示例：
```
[GROUP_ADMIN, *, projects, create]  // GROUP_ADMIN角色可以在任何域创建项目
[MEMBER, *, files, read]            // MEMBER角色可以在任何域读取文件
[user:1, group:5, projects, update] // 用户1可以在群组5中更新项目
```

通过Casbin的RBAC功能，系统支持：
1. 用户-角色分配
2. 角色-权限管理
3. 多域权限控制
4. 动态权限调整 