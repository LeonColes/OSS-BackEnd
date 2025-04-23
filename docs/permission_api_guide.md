# 权限系统API使用指南

本文档详细说明了权限系统的所有API及其用法，帮助开发者和使用者理解如何正确使用这些接口。

## 1. 权限API概览

权限系统提供以下几类API:

| 分类 | 说明 | 适用角色 |
|-----|------|---------|
| 角色管理 | 分配和撤销用户角色 | 系统管理员、群组管理员 |
| 权限管理 | 授予和撤销权限 | 系统管理员、群组管理员 |
| 群组管理 | 创建和管理群组 | 所有用户（创建）、群组管理员（管理） |
| 成员管理 | 管理群组成员 | 群组管理员 |
| 权限检查 | 验证权限 | 所有用户 |

## 2. 角色管理API

### 2.1 给用户分配角色

**请求**:
```
POST /api/oss/role/assign
```

**参数**:
```json
{
  "user_id": 123,
  "role_code": "GROUP_ADMIN",
  "domain": "group:456",
  "description": "给用户分配群组管理员角色"
}
```

**权限要求**: 系统管理员或群组管理员

**说明**: 为指定用户分配角色，domain参数指定作用域（如群组或项目）

### 2.2 撤销用户角色

**请求**:
```
POST /api/oss/role/revoke
```

**参数**:
```json
{
  "user_id": 123,
  "role_code": "GROUP_ADMIN",
  "domain": "group:456"
}
```

**权限要求**: 系统管理员或群组管理员

**说明**: 撤销用户在指定域的特定角色

## 3. 权限管理API

### 3.1 授予权限

**请求**:
```
POST /api/oss/permission/grant
```

**参数**:
```json
{
  "subject": "user:123",
  "domain": "group:456",
  "object": "file",
  "action": "upload"
}
```

**权限要求**: 群组管理员

**说明**: 授予用户对指定资源的权限

### 3.2 撤销权限

**请求**:
```
POST /api/oss/permission/revoke
```

**参数**:
```json
{
  "subject": "user:123",
  "domain": "group:456",
  "object": "file",
  "action": "delete"
}
```

**权限要求**: 群组管理员

**说明**: 撤销用户对指定资源的特定权限

### 3.3 批量授予权限

**请求**:
```
POST /api/oss/permission/batch-grant
```

**参数**:
```json
{
  "rules": [
    {
      "subject": "user:123",
      "domain": "group:456",
      "object": "file",
      "action": "upload"
    },
    {
      "subject": "user:123",
      "domain": "group:456",
      "object": "file",
      "action": "download"
    }
  ]
}
```

**权限要求**: 群组管理员

**说明**: 一次性授予多个权限

## 4. 群组管理API

### 4.1 创建群组

**请求**:
```
POST /api/oss/group/create
```

**参数**:
```json
{
  "name": "研发团队",
  "description": "技术研发小组",
  "quota": 100,
  "access_mode": "private"
}
```

**权限要求**: 任何已登录用户

**说明**: 创建新群组，创建者自动成为该群组的管理员

### 4.2 查看我的群组列表

**请求**:
```
GET /api/oss/group/my-groups
```

**权限要求**: 任何已登录用户

**说明**: 获取当前用户参与的所有群组

### 4.3 查看群组详情

**请求**:
```
GET /api/oss/group/{id}
```

**权限要求**: 群组成员

**说明**: 查看指定群组的详细信息

## 5. 成员管理API

### 5.1 邀请用户加入群组

**请求**:
```
POST /api/oss/group/{id}/invite
```

**参数**:
```json
{
  "user_ids": [123, 456],
  "message": "欢迎加入我们的团队",
  "expiry": 86400
}
```

**权限要求**: 群组管理员

**说明**: 邀请用户加入群组，可选指定邀请链接有效期（秒）

### 5.2 生成邀请码

**请求**:
```
POST /api/oss/group/{id}/invite-code
```

**参数**:
```json
{
  "expiry": 86400,
  "max_uses": 10
}
```

**权限要求**: 群组管理员

**说明**: 生成可共享的邀请码，可设置有效期和最大使用次数

### 5.3 通过邀请码加入群组

**请求**:
```
POST /api/oss/group/join
```

**参数**:
```json
{
  "invite_code": "abcd1234"
}
```

**权限要求**: 任何已登录用户

**说明**: 使用邀请码加入群组

### 5.4 移除群组成员

**请求**:
```
DELETE /api/oss/group/{id}/member/{user_id}
```

**权限要求**: 群组管理员

**说明**: 将指定用户从群组中移除

## 6. 权限检查API

### 6.1 检查用户权限

**请求**:
```
GET /api/oss/permission/check
```

**参数**:
```
?domain=group:456&object=file&action=upload
```

**权限要求**: 任何已登录用户

**说明**: 检查当前用户是否拥有特定权限

### 6.2 获取用户在域内的所有权限

**请求**:
```
GET /api/oss/permission/list
```

**参数**:
```
?domain=group:456
```

**权限要求**: 已登录用户（仅能查看自己的权限）或群组管理员

**说明**: 列出用户在指定域内的所有权限

## 7. 使用示例

### 7.1 创建群组并邀请成员

```javascript
// 1. 创建群组
const groupResponse = await fetch('/api/oss/group/create', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: '项目A研发团队',
    description: '负责项目A的研发工作',
    quota: 100
  })
});
const group = await groupResponse.json();

// 2. 生成邀请码
const inviteCodeResponse = await fetch(`/api/oss/group/${group.data.id}/invite-code`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    expiry: 86400,
    max_uses: 10
  })
});
const inviteCode = await inviteCodeResponse.json();

// 3. 分享邀请码给其他用户
console.log('邀请码:', inviteCode.data.code);
```

### 7.2 管理成员权限

```javascript
// 1. 给用户授予文件上传权限
await fetch('/api/oss/permission/grant', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    subject: `user:${userId}`,
    domain: `group:${groupId}`,
    object: 'file',
    action: 'upload'
  })
});

// 2. 给用户授予文件删除权限
await fetch('/api/oss/permission/grant', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    subject: `user:${userId}`,
    domain: `group:${groupId}`,
    object: 'file',
    action: 'delete'
  })
});

// 3. 撤销文件删除权限
await fetch('/api/oss/permission/revoke', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    subject: `user:${userId}`,
    domain: `group:${groupId}`,
    object: 'file',
    action: 'delete'
  })
});
```

## 8. 最佳实践

### 8.1 权限检查流程

在前端实现权限检查时，建议遵循以下流程：

1. **登录后缓存基本角色信息**：在用户登录后，获取并缓存用户的基本角色信息
2. **按需获取特定权限**：针对特定操作，使用权限检查API验证
3. **UI元素可见性控制**：基于权限检查结果动态调整UI元素的可见性
4. **后端双重验证**：前端的权限检查仅用于改善用户体验，所有操作在后端仍需验证权限

### 8.2 常见的权限组合

为简化权限管理，可以定义一些常用的权限组合：

| 组合名称 | 包含权限 | 适用场景 |
|---------|---------|---------|
| 只读权限 | read, download | 仅需查看和下载的用户 |
| 基本权限 | read, upload, download | 普通协作成员 |
| 高级权限 | read, upload, download, delete | 项目核心成员 |
| 完全权限 | read, upload, download, delete, share | 高级成员 |

### 8.3 权限调试

在开发过程中，可以使用以下API进行权限调试：

**请求**:
```
GET /api/oss/permission/debug
```

**参数**:
```
?subject=user:123&domain=group:456&object=file&action=upload
```

**权限要求**: 仅开发环境可用

**说明**: 返回详细的权限验证过程和结果，帮助开发者理解权限验证逻辑 