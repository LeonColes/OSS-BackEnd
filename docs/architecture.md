# OSS-Backend 系统架构设计

<div align="center">
  
## 系统架构图

```mermaid
graph TD
    User[用户] --> WebUI[Web界面]
    User --> MobileApp[移动应用]
    User --> CLI[命令行工具]
    
    subgraph "用户操作层"
      WebUI
      MobileApp
      CLI
    end
    
    WebUI --> Gateway[API网关/负载均衡]
    MobileApp --> Gateway
    CLI --> Gateway
    
    Gateway --> Monitor[监控系统\nPrometheus]
    Gateway --> Logger[日志系统\nELK/Loki]
    
    subgraph "OSS-Backend服务"
      Gateway --> UserSrv[用户服务]
      Gateway --> AuthSrv[权限服务]
      Gateway --> StorageSrv[存储服务]
      Gateway --> TaskSrv[任务服务]
      
      UserSrv --> UserMgt[用户管理]
      AuthSrv --> RBAC[RBAC权限\nCasbin]
      StorageSrv --> FileMgt[文件管理]
      TaskSrv --> TaskScheduler[任务调度]
    end
    
    subgraph "中间件层"
      UserMgt --> Redis[Redis缓存]
      RBAC --> Redis
      FileMgt --> Redis
      TaskScheduler --> Redis
      
      UserMgt --> MsgQueue[消息队列\nKafka/NATS]
      RBAC --> MsgQueue
      FileMgt --> MsgQueue
      TaskScheduler --> MsgQueue
      
      UserMgt --> Discovery[服务发现\nConsul/etcd]
      RBAC --> Discovery
      FileMgt --> Discovery
      TaskScheduler --> Discovery
    end
    
    subgraph "存储层"
      Redis --> DB[MySQL/PG\n元数据存储]
      MsgQueue --> DB
      Discovery --> DB
      
      Redis --> ObjectStore[MinIO\n对象存储]
      MsgQueue --> ObjectStore
      Discovery --> ObjectStore
    end
    
    style User fill:#f9f9f9,stroke:#333,stroke-width:2px
    style WebUI fill:#d0e0ff,stroke:#333,stroke-width:1px
    style MobileApp fill:#d0e0ff,stroke:#333,stroke-width:1px
    style CLI fill:#d0e0ff,stroke:#333,stroke-width:1px
    
    style Gateway fill:#ffe0b2,stroke:#333,stroke-width:1px
    style Monitor fill:#ffccbc,stroke:#333,stroke-width:1px
    style Logger fill:#ffccbc,stroke:#333,stroke-width:1px
    
    style UserSrv fill:#c8e6c9,stroke:#333,stroke-width:1px
    style AuthSrv fill:#c8e6c9,stroke:#333,stroke-width:1px
    style StorageSrv fill:#c8e6c9,stroke:#333,stroke-width:1px
    style TaskSrv fill:#c8e6c9,stroke:#333,stroke-width:1px
    
    style UserMgt fill:#b3e5fc,stroke:#333,stroke-width:1px
    style RBAC fill:#b3e5fc,stroke:#333,stroke-width:1px
    style FileMgt fill:#b3e5fc,stroke:#333,stroke-width:1px
    style TaskScheduler fill:#b3e5fc,stroke:#333,stroke-width:1px
    
    style Redis fill:#e1bee7,stroke:#333,stroke-width:1px
    style MsgQueue fill:#e1bee7,stroke:#333,stroke-width:1px
    style Discovery fill:#e1bee7,stroke:#333,stroke-width:1px
    
    style DB fill:#bbdefb,stroke:#333,stroke-width:1px
    style ObjectStore fill:#bbdefb,stroke:#333,stroke-width:1px
```

</div>

> **系统架构总览**: OSS-Backend是一个完整的对象存储服务，采用微服务架构，提供高性能、安全可靠的文件存储与管理功能

## 📋 目录

- [1. 系统概述](#1-系统概述)
- [2. 架构设计原则](#2-架构设计原则)
- [3. 整体架构](#3-整体架构)
- [4. 技术栈选型](#4-技术栈选型)
- [5. 核心模块设计](#5-核心模块设计)
- [6. 存储设计](#6-存储设计)
- [7. 认证与授权设计](#7-认证与授权设计)
- [8. 部署架构](#8-部署架构)
- [9. 性能与扩展性](#9-性能与扩展性)
- [10. 安全设计](#10-安全设计)

---

## 1. 系统概述

> OSS-Backend是一个基于Golang开发的对象存储服务后端系统，提供文件上传、下载、管理和访问控制等功能。系统设计采用微服务架构思想，将不同功能模块解耦，提高系统的可维护性和扩展性。

### 💠 核心层次结构

| 层级 | 说明 | 主要组件 |
|------|------|---------|
| **用户操作层** | 包括各种用户交互界面 | Web界面、移动应用、命令行工具 |
| **API网关层** | 统一入口，请求路由 | 负载均衡、认证鉴权组件 |
| **服务层** | 核心业务逻辑实现 | 用户服务、权限服务、存储服务、任务服务 |
| **中间件层** | 提供基础设施支持 | Redis缓存、消息队列、服务发现 |
| **存储层** | 负责数据持久化 | 元数据存储(MySQL/PG)、对象存储(MinIO) |

### 🔄 用户操作流程

<div align="center">

```mermaid
sequenceDiagram
    actor User as 用户
    participant Web as Web/移动界面
    participant API as API网关
    participant Auth as 认证服务
    participant Project as 项目服务
    participant Storage as 存储服务
    participant Minio as MinIO
    
    User->>Web: 1. 注册/登录
    Web->>API: 发送认证请求
    API->>Auth: 验证凭证
    Auth-->>API: 返回JWT令牌
    API-->>Web: 返回认证结果
    
    User->>Web: 2. 创建项目
    Web->>API: 发送创建项目请求
    API->>Project: 创建新项目
    Project-->>API: 返回项目信息
    API-->>Web: 返回创建结果
    
    User->>Web: 3. 上传文件
    Web->>API: 发送上传请求
    API->>Storage: 处理文件上传
    Storage->>Minio: 存储文件数据
    Minio-->>Storage: 确认存储成功
    Storage-->>API: 返回文件元数据
    API-->>Web: 返回上传结果
    
    User->>Web: 4. 管理权限
    Web->>API: 发送权限设置请求
    API->>Auth: 更新资源权限
    Auth-->>API: 确认权限更新
    API-->>Web: 返回设置结果
    
    User->>Web: 5. 文件操作
    Web->>API: 发送文件操作请求
    API->>Storage: 执行文件操作
    Storage->>Minio: 访问文件数据
    Minio-->>Storage: 返回文件数据
    Storage-->>API: 处理完成响应
    API-->>Web: 返回操作结果
```

</div>

1. **🔐 用户注册/登录**: 用户通过Web界面或移动应用注册账号并登录系统
2. **📂 项目创建**: 用户创建项目作为文件组织的容器
3. **📤 文件上传**: 用户将文件上传到指定项目，系统处理文件并存储
4. **🔒 权限设置**: 用户可设置文件/项目的访问权限，如私有、共享或公开
5. **🔧 文件管理**: 用户可进行文件查看、下载、删除、重命名等操作
6. **🔄 版本控制**: 系统支持文件版本控制，可查看和恢复历史版本

---

## 2. 架构设计原则

<div align="center">
<table>
  <tr>
    <td align="center"><h3>📐</h3><strong>领域驱动设计</strong><br/><small>基于业务领域构建系统架构</small></td>
    <td align="center"><h3>🧩</h3><strong>整洁架构</strong><br/><small>关注点分离，依赖由外向内</small></td>
    <td align="center"><h3>🔌</h3><strong>微服务架构</strong><br/><small>服务解耦，独立部署和扩展</small></td>
  </tr>
  <tr>
    <td align="center"><h3>🔒</h3><strong>安全第一</strong><br/><small>数据安全和访问控制贯穿设计始终</small></td>
    <td align="center"><h3>⚖️</h3><strong>可扩展性</strong><br/><small>支持水平扩展以应对业务增长</small></td>
    <td align="center"><h3>📊</h3><strong>可观测性</strong><br/><small>内置监控、日志和追踪能力</small></td>
  </tr>
</table>
</div>

---

## 3. 整体架构

> 系统采用分层架构设计，实现了关注点分离和责任清晰化

<div align="center">

```mermaid
classDiagram
    class InterfaceLayer {
        HTTP API
        gRPC
        WebSocket
        GraphQL
    }
    
    class ApplicationLayer {
        服务编排
        用例实现
        事务管理
    }
    
    class DomainLayer {
        业务实体
        值对象
        领域服务
        聚合
    }
    
    class InfrastructureLayer {
        数据库访问
        第三方集成
        消息队列
        缓存等
    }
    
    InterfaceLayer --> ApplicationLayer
    ApplicationLayer --> DomainLayer
    DomainLayer --> InfrastructureLayer
```

</div>

### 🏢 核心服务组件

<div align="center">
<img src="https://via.placeholder.com/800x400.png?text=OSS-Backend+核心服务组件" alt="核心服务组件" style="max-width:80%;">
</div>

- **🌐 API网关**: 统一入口，请求路由，认证鉴权
- **👤 用户服务**: 用户管理，身份认证
- **🔑 权限服务**: 基于RBAC+Casbin的权限控制
- **💾 存储服务**: 文件存储管理，包含元数据和数据存储
- **⏱️ 任务调度服务**: 异步任务处理
- **📢 通知服务**: 系统通知和消息推送
- **📊 监控服务**: 系统监控和日志收集

---

## 4. 技术栈选型

### 🚀 编程语言与框架

<div align="center">
<table>
  <tr>
    <th>类别</th>
    <th>技术选择</th>
    <th>说明</th>
  </tr>
  <tr>
    <td>主语言</td>
    <td><strong>Go 1.21+</strong></td>
    <td>高性能、低资源占用、并发友好</td>
  </tr>
  <tr>
    <td>Web框架</td>
    <td><strong>Gin</strong></td>
    <td>轻量、高性能的HTTP Web框架</td>
  </tr>
  <tr>
    <td>RPC框架</td>
    <td><strong>gRPC</strong></td>
    <td>高性能、跨语言的RPC框架</td>
  </tr>
  <tr>
    <td>API文档</td>
    <td><strong>Swagger/OpenAPI</strong></td>
    <td>RESTful API的设计和文档工具</td>
  </tr>
</table>
</div>

### 💾 存储层

- **关系型数据库**: PostgreSQL (元数据存储)
- **对象存储**: MinIO (文件数据存储)
- **缓存**: Redis
- **搜索引擎**: Elasticsearch (可选)

### 🔧 中间件与基础设施

- **消息队列**: Kafka/NATS
- **服务发现**: Consul/etcd
- **日志收集**: ELK/Loki
- **监控系统**: Prometheus + Grafana
- **链路追踪**: Jaeger/Zipkin

### 🚢 部署与运维

- **容器化**: Docker
- **编排系统**: Kubernetes
- **CI/CD**: GitHub Actions/Jenkins
- **配置管理**: Helm

---

## 5. 核心模块设计

### 👤 用户管理模块

<div align="center">

```mermaid
flowchart LR
    UI[用户接口] --> AS[用户应用服务]
    AS --> DM[用户领域]
    AS --> UR[用户资源库]
    DM --> UR
    UR <--> US[用户存储]
    
    style UI fill:#f9f0ff,stroke:#333,stroke-width:1px
    style AS fill:#e0f7fa,stroke:#333,stroke-width:1px
    style DM fill:#e8f5e9,stroke:#333,stroke-width:1px
    style UR fill:#fff3e0,stroke:#333,stroke-width:1px
    style US fill:#f3e5f5,stroke:#333,stroke-width:1px
```

</div>

提供用户注册、登录、个人信息管理、认证等功能，包括：

- 多种认证方式支持（账密、OAuth、LDAP等）
- 用户信息管理
- 安全设置与MFA
- 用户组管理

### 🔑 权限管理模块

<div align="center">

```mermaid
graph TD
    User[用户] --> Role[角色]
    Group[用户组] --> Role
    Role --> Permission[权限]
    Permission --> Resource[资源]
    
    style User fill:#bbdefb,stroke:#333,stroke-width:1px
    style Group fill:#bbdefb,stroke:#333,stroke-width:1px
    style Role fill:#c8e6c9,stroke:#333,stroke-width:1px
    style Permission fill:#ffe0b2,stroke:#333,stroke-width:1px
    style Resource fill:#ffccbc,stroke:#333,stroke-width:1px
```

</div>

基于RBAC模型和Casbin实现的动态权限系统，支持多维度的访问控制：

- 角色定义与管理
- 权限分配与继承
- 资源ACL控制
- API级别权限验证
- 数据行级权限控制

### 💾 文件存储模块

<div align="center">

```mermaid
flowchart LR
    FI[文件操作接口] --> FS[文件应用服务]
    FS --> FD[文件领域]
    FD --> FM[文件元数据存储]
    FD --> FDS[文件数据存储]
    
    style FI fill:#f9f0ff,stroke:#333,stroke-width:1px
    style FS fill:#e0f7fa,stroke:#333,stroke-width:1px
    style FD fill:#e8f5e9,stroke:#333,stroke-width:1px
    style FM fill:#fff3e0,stroke:#333,stroke-width:1px
    style FDS fill:#f3e5f5,stroke:#333,stroke-width:1px
```

</div>

负责文件的上传、下载和管理：

- 大文件分片上传
- 断点续传
- 文件版本控制
- 元数据管理
- 文件加密存储
- 数据去重

### ⏱️ 任务调度模块

处理异步任务和长时间运行的作业：

- 文件处理（压缩、格式转换等）
- 批量操作
- 定时任务
- 重试机制
- 分布式作业调度

---

## 6. 存储设计

### 🗃️ 元数据存储

<div align="center">

```mermaid
erDiagram
    users ||--o{ user_roles : has
    users ||--o{ groups : belongs_to
    roles ||--o{ user_roles : assigned_to
    roles ||--o{ role_permissions : has
    permissions ||--o{ role_permissions : assigned_to
    files ||--o{ file_versions : has
    files ||--o{ file_tags : has
    tags ||--o{ file_tags : assigned_to
    
    users {
        uuid id PK
        string username
        string email
        string password_hash
        datetime created_at
        datetime updated_at
    }
    
    files {
        uuid id PK
        string filename
        string path
        string content_type
        int64 size
        uuid owner_id FK
        datetime created_at
        datetime updated_at
    }
```

</div>

使用PostgreSQL存储系统元数据：

- 用户信息
- 权限配置
- 文件元数据
- 系统配置

### 📁 文件数据存储

<div align="center">
<img src="https://via.placeholder.com/800x300.png?text=MinIO对象存储架构" alt="MinIO对象存储架构" style="max-width:80%;">
</div>

使用MinIO作为对象存储后端：

- 按租户隔离存储桶
- 分层存储策略
- 内容寻址存储
- 加密存储支持

---

## 7. 认证与授权设计

### 🔐 认证流程

<div align="center">

```mermaid
sequenceDiagram
    actor User as 用户
    participant Client as 客户端
    participant API as API网关
    participant Auth as 认证服务
    participant DB as 用户数据库
    
    User->>Client: 输入凭证
    Client->>API: 发送认证请求
    API->>Auth: 转发认证请求
    Auth->>DB: 验证凭证
    DB-->>Auth: 返回用户信息
    Auth-->>API: 生成JWT令牌
    API-->>Client: 返回令牌
    Client->>API: 带令牌请求资源
    API->>Auth: 验证令牌
    Auth-->>API: 认证通过
    API-->>Client: 返回资源
```

</div>

1. **多因素认证**: 支持密码、令牌、证书等多种认证方式
2. **JWT令牌**: 无状态会话管理
3. **OAuth集成**: 支持第三方登录
4. **会话管理**: 登录状态控制与安全退出

### 🔒 授权模型

<div class="authorization-model">
<pre>
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
</pre>
</div>

---

## 8. 部署架构

### 🖥️ 单体部署

<div align="center">

```mermaid
flowchart TD
    OSS[OSS-Backend] --> DB[(PostgreSQL/Redis)]
    DB --> MinIO[(MinIO)]
    
    style OSS fill:#bbdefb,stroke:#333,stroke-width:2px
    style DB fill:#c8e6c9,stroke:#333,stroke-width:1px
    style MinIO fill:#ffe0b2,stroke:#333,stroke-width:1px
```

</div>

### 🌐 微服务部署

<div align="center">

```mermaid
flowchart TD
    API[API Gateway] --> US[用户服务]
    API --> AS[权限服务]
    API --> SS[存储服务]
    API --> TS[任务服务]
    
    US --> DB[(Shared DB/Cache)]
    AS --> DB
    SS --> DB
    TS --> DB
    
    DB --> OS[(Object Storage)]
    
    style API fill:#bbdefb,stroke:#333,stroke-width:2px
    style US fill:#c8e6c9,stroke:#333,stroke-width:1px
    style AS fill:#c8e6c9,stroke:#333,stroke-width:1px
    style SS fill:#c8e6c9,stroke:#333,stroke-width:1px
    style TS fill:#c8e6c9,stroke:#333,stroke-width:1px
    style DB fill:#ffe0b2,stroke:#333,stroke-width:1px
    style OS fill:#ffccbc,stroke:#333,stroke-width:1px
```

</div>

---

## 9. 性能与扩展性

### ⚡ 性能优化策略

<div align="center">
<table>
  <tr>
    <td align="center"><h3>📊</h3><strong>连接池管理</strong><br/><small>优化数据库连接</small></td>
    <td align="center"><h3>⚡</h3><strong>缓存策略</strong><br/><small>多级缓存减少I/O</small></td>
    <td align="center"><h3>🔄</h3><strong>异步处理</strong><br/><small>非关键流程异步化</small></td>
  </tr>
  <tr>
    <td align="center"><h3>🚦</h3><strong>限流与降级</strong><br/><small>保护系统稳定性</small></td>
    <td align="center"><h3>🔥</h3><strong>预热与预取</strong><br/><small>减少冷启动开销</small></td>
    <td align="center"><h3>📈</h3><strong>监控与调优</strong><br/><small>持续性能优化</small></td>
  </tr>
</table>
</div>

### 📈 扩展性设计

- **水平扩展**: 无状态设计支持集群扩展
- **分片策略**: 按租户/时间分片数据
- **容量规划**: 弹性资源分配
- **热点识别**: 动态调整热点资源

---

## 10. 安全设计

### 🔒 数据安全

<div align="center">
<img src="https://via.placeholder.com/800x300.png?text=OSS-Backend数据安全架构" alt="数据安全架构" style="max-width:80%;">
</div>

- **传输加密**: TLS/SSL通信加密
- **存储加密**: 文件加密存储
- **密钥管理**: KMS密钥统一管理
- **数据脱敏**: 敏感信息脱敏展示

### 🛡️ 应用安全

- **请求验证**: 输入数据验证
- **CSRF防护**: 跨站请求伪造防护
- **XSS防御**: 跨站脚本攻击防御
- **权限检查**: 多层次权限校验
- **日志审计**: 关键操作审计追踪

---

<div align="center">
<strong>该文档将随系统发展持续更新，所有重大架构变更需经过架构评审并更新本文档。</strong>
</div> 