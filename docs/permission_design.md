# OSS-Backend 权限系统设计文档

## 1. 系统概述

本系统采用基于 Go + Gin + Casbin + GORM 实现的动态 RBAC 权限控制框架，结合多租户（域）模型，实现了系统级、群组级和项目级三层权限隔离。系统支持角色动态分配，权限实时校验，确保不同级别的用户只能访问其授权范围内的资源。

### 1.1 技术选型

- **Casbin**: 强大的跨语言访问控制框架，支持多种权限模型
- **GORM**: Go语言ORM库，用于数据持久化
- **Gin**: 高性能Web框架，用于API接口实现
- **RBAC with Domains**: 基于角色的访问控制，增加域隔离

### 1.2 核心特性

- 三层权限控制：系统级、群组级、项目级
- 多租户隔离：不同群组/项目的权限互不影响
- 动态权限分配：支持运行时修改权限策略
- 统一中间件校验：所有API请求统一进行权限校验

## 2. 角色设计

### 2.1 角色体系

系统定义了三个核心角色，分别负责不同层级的权限管理：

| 角色名称 | 角色编码 | 职责概述 |
|---------|----------|---------|
| 系统管理员 | SUPER_ADMIN | 拥有系统全局权限，可管理所有组织和资源 |
| 群组管理员 | GROUP_ADMIN | 负责管理群组和项目，控制项目数量和资源分配 |
| 项目管理员 | PROJECT_ADMIN | 负责管理项目成员，控制项目人数和成员权限 |
| 普通成员 | MEMBER | 具有基本的文件操作权限，可以参与项目协作 |

### 2.2 角色权限矩阵

| 功能/操作 | 系统管理员 | 群组管理员 | 项目管理员 | 普通成员 |
|----------|----------|-----------|-----------|---------|
| **用户管理** |  |  |  |  |
| 查看个人信息 | ✓ | ✓ | ✓ | ✓ |
| 修改个人信息 | ✓ | ✓ | ✓ | ✓ |
| 查看用户列表 | ✓ | ✓ | ✓ | ✗ |
| 查看群组列表 | ✓ | ✓ | ✗ | ✗ |
| **项目管理** |  |  |  |  |
| 创建项目 | ✓ | ✓ | ✗ | ✗ |
| 编辑项目信息 | ✓ | ✓ | ✓ | ✗ |
| 查看所有项目 | ✓ | ✓ | ✗ | ✗ |
| 查看参与项目 | ✓ | ✓ | ✓ | ✓ |
| 删除项目(逻辑) | ✓ | ✓ | ✗ | ✗ |
| **成员管理** |  |  |  |  |
| 添加项目成员 | ✓ | ✓ | ✓ | ✗ |
| 移除项目成员 | ✓ | ✓ | ✓ | ✗ |
| 分配角色 | ✓ | ✓ | ✓ | ✗ |
| **文件管理** |  |  |  |  |
| 上传文件 | ✓ | ✓ | ✓ | ✓ |
| 下载文件 | ✓ | ✓ | ✓ | ✓ |
| 删除文件(逻辑) | ✓ | ✓ | ✓ | ✓ |
| 查看文件列表 | ✓ | ✓ | ✓ | ✓ |

## 3. 权限实现机制

### 3.1 Casbin模型设计

系统采用的RBAC with Domains模型定义:

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && (r.obj == p.obj || p.obj == '*') && (r.act == p.act || p.act == '*')
```

- **sub**: 主体 - 用户或角色
- **dom**: 域 - system/group/project
- **obj**: 对象 - 资源路径
- **act**: 行为 - 操作类型（read/create/update/delete）

### 3.2 权限层级结构

1. **系统级权限**
   - 域: `system`
   - 对象: API路径，如 `/api/oss/group/*`
   - 行为: HTTP方法或抽象操作

2. **群组级权限**
   - 域: `group:{groupID}`
   - 对象: 资源类型，如 `file`、`member`
   - 行为: 抽象操作，如 `read`、`create`

3. **项目级权限**
   - 域: `project:{projectID}`
   - 对象: 资源类型，如 `file`、`member`
   - 行为: 抽象操作，如 `read`、`create`

### 3.3 数据模型

```go
// CasbinRule Casbin规则实体
type CasbinRule struct {
    ID    uint   `gorm:"primaryKey;autoIncrement"`
    Ptype string `gorm:"size:100;uniqueIndex:unique_policy;comment:策略类型"`
    V0    string `gorm:"size:100;uniqueIndex:unique_policy;comment:角色/用户"`
    V1    string `gorm:"size:100;uniqueIndex:unique_policy;comment:域/租户"`
    V2    string `gorm:"size:100;uniqueIndex:unique_policy;comment:资源"`
    V3    string `gorm:"size:100;uniqueIndex:unique_policy;comment:操作"`
    V4    string `gorm:"size:100;uniqueIndex:unique_policy;comment:预留字段"`
    V5    string `gorm:"size:100;uniqueIndex:unique_policy;comment:预留字段"`
}
```

## 4. 权限验证流程

### 4.1 API接口权限验证

所有API请求通过中间件链进行权限验证：

```
客户端请求 -> JWT认证 -> 权限校验中间件 -> 业务处理 -> 响应
```

### 4.2 中间件实现

系统针对不同级别的资源，实现了三种权限校验中间件：

1. **系统级权限中间件** (`AuthCheckForSystem`):
   - 校验用户对系统级API的访问权限
   - 基于角色和API路径进行验证

2. **群组级权限中间件** (`AuthCheckForGroup`):
   - 识别请求中的群组ID
   - 构造特定域的权限验证请求
   - 校验用户对群组资源的操作权限

3. **项目级权限中间件** (`AuthCheckForProject`):
   - 识别请求中的项目ID
   - 构造特定域的权限验证请求
   - 校验用户对项目资源的操作权限

### 4.3 权限判断逻辑

1. **HTTP方法映射**:
   ```go
   func mapMethodToAction(method string) string {
       switch method {
       case http.MethodGet:
           return "read"
       case http.MethodPost:
           return "create"
       case http.MethodPut:
           return "update"
       case http.MethodDelete:
           return "delete"
       default:
           return "*"
       }
   }
   ```

2. **权限检查**:
   - 首先检查用户直接权限
   - 然后检查用户所属角色的权限
   - 任一权限满足即允许访问

## 5. 权限相关API

系统提供以下API用于权限管理：

| 方法 | 接口路径 | 角色要求 | 说明 |
|-----|---------|----------|-----|
| POST | /api/oss/role/assign | SUPER_ADMIN 或 GROUP_ADMIN | 给用户分配角色 |
| POST | /api/oss/permission/grant | SUPER_ADMIN | 给角色授予权限 |
| POST | /api/oss/permission/revoke | SUPER_ADMIN | 撤销角色权限 |
| POST | /api/oss/group/admin/set | SUPER_ADMIN | 设置群组管理员 |
| POST | /api/oss/project/member/add | GROUP_ADMIN 或 PROJECT_ADMIN | 添加项目成员 |
| GET | /api/oss/permission/check | 任意角色 | 检查用户权限 |

## 6. 角色分配流程

### 6.1 新用户角色分配

1. 用户注册时自动分配MEMBER角色
2. 系统管理员可将用户提升为GROUP_ADMIN
3. 群组管理员可将成员提升为PROJECT_ADMIN

### 6.2 角色升级流程

项目成员升级为项目管理员：
```
群组管理员 -> 查看项目成员 -> 选择成员 -> 分配PROJECT_ADMIN角色 -> 确认
```

### 6.3 权限继承

1. 系统管理员(SUPER_ADMIN)拥有全部权限
2. 群组管理员(GROUP_ADMIN)拥有其管理的群组全部权限
3. 项目管理员(PROJECT_ADMIN)拥有其管理的项目全部权限

## 7. 优化建议

为进一步提升权限系统的效率和易用性，建议进行以下优化：

1. **中间件整合**：
   - 将三个权限中间件整合为一个统一的权限检查中间件
   - 通过参数配置指定检查的权限级别

2. **缓存优化**：
   - 使用Redis缓存常用权限策略
   - 实现策略变更时的缓存自动失效

3. **角色命名优化**：
   - 将角色名称调整为更直观的命名：
     - SUPER_ADMIN → system_admin
     - GROUP_ADMIN → org_admin
     - PROJECT_ADMIN、MEMBER → project_admin、project_member

4. **权限可视化**：
   - 开发权限管理界面，直观展示和编辑权限配置

## 8. 总结

OSS-Backend的权限系统采用RBAC with Domains模型，实现了系统级、群组级和项目级三层权限控制，确保各级资源的访问安全。通过Casbin框架实现权限的动态管理和实时校验，满足企业级对象存储系统的复杂权限需求。 