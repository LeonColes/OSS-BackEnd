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
| **/api/oss/role/create** | ✓ | ✓ | ✗ | 创建角色（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/update** | ✓ | ✓ | ✗ | 更新角色（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/delete/:id** | ✓ | ✓ | ✗ | 删除角色（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/detail/:id** | ✓ | ✓ | ✗ | 角色详情（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/role/list** | ✓ | ✓ | ✗ | 角色列表（需要ADMIN或GROUP_ADMIN权限） |
| **/api/oss/group/create** | ✓ | ✓ | ✓ | 创建群组（需登录） |

## 角色权限详细说明

3. **角色权限详细说明**：
   - **ADMIN**：拥有所有系统权限
   - **GROUP_ADMIN**：
     - 可创建/管理项目(projects的create/read/update/delete)
     - 可读取/更新群组(groups的read/update)
     - 可读取用户信息(users的read)
     - 可分配角色(roles的assign)
     - 可管理成员(members的add/remove)
     - 可对文件进行所有操作(files的create/read/update/delete)
     - 可创建和管理自定义角色(roles的create/read/update/delete)
   - **MEMBER**：
     - 可读取项目(projects的read)
     - 可对文件进行所有操作(files的create/read/update/delete)

// ... existing code ... 