# 后端设计文档

## 一、技术架构

### 1. 技术栈选择
- **编程语言**：Golang
- **Web框架**：Gin
- **数据库**：MySQL 8.0+
- **缓存**：Redis 6.0+
- **对象存储**：MinIO
- **API文档**：Swagger
- **日志**：Zap
- **配置管理**：Viper

### 2. 系统架构
采用分层架构设计：
- **API层**：处理HTTP请求，参数校验，权限检查
- **Service层**：实现业务逻辑
- **Repository层**：数据访问抽象
- **Model层**：数据模型定义
- **Util层**：工具函数

```
oss-backend/
├── main.go                          # 应用主入口
├── go.mod                           # Go 模块文件
├── go.sum                           # Go 模块文件
├── README.md                        # 项目说明
├── docs/                            # 文档目录及API文档
│   └── swagger/                     # Swagger自动生成文档
├── configs/                         # 配置文件目录
├── internal/                        # 内部代码
│   ├── model/                       # 数据模型
│   │   ├── entity/                  # 数据库实体
│   │   │   ├── role.go              # 角色实体
│   │   │   └── user.go              # 用户实体
│   │   └── dto/                     # 数据传输对象
│   │       ├── role_dto.go          # 角色DTO
│   │       └── user_dto.go          # 用户DTO
│   ├── controller/                  # 控制器层
│   │   ├── role_controller.go       # 角色控制器
│   │   └── user_controller.go       # 用户控制器
│   ├── service/                     # 服务层
│   │   ├── role_service.go          # 角色服务
│   │   └── user_service.go          # 用户服务
│   ├── repository/                  # 数据访问层
│   │   ├── role_repository.go       # 角色仓库
│   │   └── user_repository.go       # 用户仓库
│   └── middleware/                  # 中间件
│       └── auth.go                  # 认证中间件
├── pkg/                             # 公共代码
│   ├── common/                      # 通用工具
│   │   └── response.go              # 统一响应格式
│   └── minio/                       # 文件存储
├── routes/                          # 路由定义
│   └── routes.go                    # 所有路由
└── uploads/                         # 文件上传目录
```

### 3. 数据流架构
```
客户端 → API网关 → 认证服务 → 业务服务 → 数据存储层(MySQL/Redis/MinIO)
```

## 二、核心模块设计

### 1. 认证与授权模块
- JWT认证实现
- 访问令牌与刷新令牌机制
- 权限检查中间件
- 角色与权限管理

### 2. 存储隔离实现
- 群组与MinIO桶一对一映射
- 项目通过路径前缀实现隔离
- 访问控制通过预签名URL实现
- 隔离策略的安全验证

### 3. 文件处理模块
- 文件元数据管理
- 文件去重与秒传实现
- 分片上传管理
- 版本控制实现
- 回收站机制（所有删除均为逻辑删除）

### 4. 缓存策略
- Token缓存
- 用户权限缓存
- 热点文件元数据缓存
- 上传状态缓存

### 5. 日志与审计
- 操作日志记录
- 审计日志查询
- 异常操作检测
- 日志导出功能

## 三、数据库设计

### 1. 核心表字段设计

#### 用户表(users)
- **id**: 主键 ID
- **email**: 用户邮箱，登录凭证
- **name**: 用户姓名
- **password_hash**: 密码哈希值
- **avatar**: 用户头像URL
- **created_at**: 创建时间
- **updated_at**: 更新时间
- **last_login_at**: 最后登录时间
- **last_login_ip**: 最后登录IP
- **status**: 状态（1-正常，2-禁用，3-锁定）

#### 群组表(groups)
- **id**: 主键 ID
- **name**: 群组名称
- **description**: 群组描述
- **group_key**: MinIO桶名（唯一）
- **invite_code**: 邀请码（唯一）
- **invite_expires_at**: 邀请码过期时间
- **storage_quota**: 存储容量配额(0表示无限制)
- **creator_id**: 创建者用户ID
- **created_at**: 创建时间
- **updated_at**: 更新时间
- **status**: 状态（1-正常，2-禁用，3-锁定）

#### 群组成员表(group_members)
- **id**: 主键 ID
- **group_id**: 群组ID
- **user_id**: 用户ID
- **role**: 角色（group_owner, group_admin, member）
- **joined_at**: 加入时间
- **updated_at**: 更新时间
- **last_active_at**: 最后活跃时间
- **is_deleted**: 是否已删除（逻辑删除）

#### 项目表(projects)
- **id**: 主键 ID
- **group_id**: 所属群组ID
- **name**: 项目名称
- **description**: 项目描述
- **path_prefix**: 存储路径前缀
- **creator_id**: 创建者用户ID
- **created_at**: 创建时间
- **updated_at**: 更新时间
- **status**: 状态（1-正常，2-归档，3-删除）
- **is_deleted**: 是否已删除（逻辑删除）

#### 项目权限表(permissions)
- **id**: 主键 ID
- **user_id**: 用户ID
- **project_id**: 项目ID
- **role**: 角色（admin, uploader, reader）
- **created_at**: 创建时间
- **updated_at**: 更新时间
- **expire_at**: 权限过期时间
- **granted_by**: 授权人用户ID
- **is_deleted**: 是否已删除（逻辑删除）

#### 文件表(files)
- **id**: 主键 ID
- **project_id**: 所属项目ID
- **file_name**: 文件名
- **file_path**: 文件路径
- **full_path**: 完整路径（用于唯一性校验）
- **file_hash**: 文件哈希值
- **file_size**: 文件大小(字节)
- **mime_type**: 文件MIME类型
- **extension**: 文件扩展名
- **is_folder**: 是否为文件夹
- **is_deleted**: 是否已删除(逻辑删除)
- **uploader_id**: 上传者用户ID
- **created_at**: 创建时间
- **updated_at**: 更新时间
- **deleted_at**: 删除时间
- **deleted_by**: 删除人用户ID
- **current_version**: 当前版本号
- **preview_url**: 预览URL

#### 文件版本表(file_versions)
- **id**: 主键 ID
- **file_id**: 关联的文件ID
- **version**: 版本号
- **file_hash**: 此版本的文件哈希值
- **file_size**: 此版本的文件大小
- **uploader_id**: 此版本的上传者ID
- **created_at**: 创建时间
- **comment**: 版本说明
- **is_deleted**: 是否已删除（逻辑删除）

#### 文件分享表(file_shares)
- **id**: 主键 ID
- **file_id**: 被分享的文件ID
- **user_id**: 分享创建者ID
- **share_code**: 分享码（唯一）
- **password**: 访问密码
- **expire_at**: 过期时间
- **download_limit**: 下载次数限制（0为无限制）
- **download_count**: 已下载次数
- **created_at**: 创建时间
- **is_deleted**: 是否已删除（逻辑删除）

#### 操作日志表(logs)
- **id**: 主键 ID
- **user_id**: 操作用户ID
- **group_id**: 群组ID
- **project_id**: 项目ID
- **file_id**: 相关文件ID
- **operation**: 操作类型
- **ip_address**: 操作IP地址
- **user_agent**: 用户代理
- **status**: 操作状态码
- **created_at**: 创建时间
- **request_details**: 请求详情
- **response_details**: 响应详情
- **execution_time**: 执行时间(ms)

#### 存储统计表(storage_stats)
- **id**: 主键 ID
- **group_id**: 群组ID
- **project_id**: 项目ID
- **stat_date**: 统计日期
- **file_count**: 文件数量
- **total_size**: 总存储大小
- **increase_size**: 当日增加大小
- **created_at**: 创建时间

### 2. 索引优化
- 针对高频查询字段建立索引
- 组合索引优化复杂查询
- 按时间范围分区的日志表

### 3. 数据关系

#### ER图
```
┌───────────┐       ┌───────────┐       ┌───────────┐
│   Users   │       │  Groups   │       │ Projects  │
├───────────┤       ├───────────┤       ├───────────┤
│ id        │       │ id        │       │ id        │
│ email     │       │ name      │       │ group_id  │◄─────┐
│ name      │       │ group_key │       │ name      │      │
│ password  │       │ invite_code│      │ path_prefix│      │
└───────────┘       │ creator_id│◄──────┘ creator_id│◄─┐   │
      ▲             └───────────┘       └───────────┘  │   │
      │                    ▲                           │   │
      │                    │                           │   │
      │             ┌───────────┐                      │   │
      └─────────────┤Group_Members│◄────────────────────┘   │
      │             ├───────────┤                          │
      │             │ id        │       ┌───────────┐      │
      │             │ group_id  │◄──────┤Permissions│      │
      │             │ user_id   │       ├───────────┤      │
      │             │ role      │       │ id        │      │
      │             └───────────┘       │ user_id   │◄─────┘
      │                                 │ project_id│◄──────┐
      │                                 │ role      │       │
      │                                 └───────────┘       │
      │                                                     │
      │                                                     │
      │             ┌───────────┐       ┌───────────┐       │
      └─────────────┤   Logs    │       │   Files   │       │
                    ├───────────┤       ├───────────┤       │
                    │ id        │       │ id        │       │
                    │ user_id   │◄──────┤ project_id│◄──────┘
                    │ group_id  │       │ file_name │
                    │ project_id│◄──────┤ file_path │
                    │ file_id   │◄──────┤ file_hash │
                    │ operation │       │ file_size │
                    │ ip_address│       │ is_folder │
                    └───────────┘       │ is_deleted│
                                        └───────────┘
                                              ▲
                                              │
                                        ┌───────────┐     ┌───────────┐
                                        │File_Versions│    │File_Shares │
                                        ├───────────┤     ├───────────┤
                                        │ id        │     │ id        │
                                        │ file_id   │◄────┤ file_id   │
                                        │ version   │     │ user_id   │
                                        │ file_hash │     │ share_code│
                                        │ file_size │     │ password  │
                                        │ uploader_id│     │ expire_at │
                                        └───────────┘     └───────────┘
```

## 四、API设计

### 1. API规范
- 只使用GET和POST方法
- GET请求参数放在URL末尾
- 统一前缀：`/api/oss`
- 统一响应格式：`{code, message, data}`
- 所有删除操作均为逻辑删除

### 2. 接口详细设计

#### 认证接口

| 方法 | 路径 | 描述 | 请求参数 | 响应 | 权限要求 |
|------|------|------|---------|------|----------|
| POST | /api/oss/auth/register | 用户注册 | ```{ email: string, password: string, name: string }``` | ```{ code: int, message: string, data: { userId: int } }``` | 无 |
| POST | /api/oss/auth/login | 用户登录 | ```{ email: string, password: string }``` | ```{ code: int, message: string, data: { token: string, refreshToken: string, userInfo: object } }``` | 无 |
| GET | /api/oss/auth/user | 获取用户信息 | 无 | ```{ code: int, message: string, data: { userInfo: object } }``` | 已登录 |
| POST | /api/oss/auth/refresh | 刷新Token | ```{ refreshToken: string }``` | ```{ code: int, message: string, data: { token: string, refreshToken: string } }``` | 无 |
| POST | /api/oss/auth/logout | 注销登录 | 无 | ```{ code: int, message: string }``` | 已登录 |
| POST | /api/oss/auth/password | 修改密码 | ```{ oldPassword: string, newPassword: string }``` | ```{ code: int, message: string }``` | 已登录 |

#### 群组管理接口

| 方法 | 路径 | 描述 | 请求参数 | 响应 | 权限要求 |
|------|------|------|---------|------|----------|
| POST | /api/oss/group/create | 创建群组 | ```{ name: string, description: string }``` | ```{ code: int, message: string, data: { groupId: int, groupKey: string } }``` | 已登录 |
| POST | /api/oss/group/join | 加入群组 | ```{ inviteCode: string }``` | ```{ code: int, message: string, data: { groupId: int } }``` | 已登录 |
| GET | /api/oss/group/list?page=1&size=10 | 获取群组列表 | 无 | ```{ code: int, message: string, data: { groups: array, total: int } }``` | 已登录 |
| GET | /api/oss/group/members?groupId=1&page=1&size=10 | 获取群组成员 | 无 | ```{ code: int, message: string, data: { members: array, total: int } }``` | 群组成员 |
| POST | /api/oss/group/member/manage | 管理成员 | ```{ groupId: int, userId: int, role: string }``` | ```{ code: int, message: string }``` | 群组管理员 |
| POST | /api/oss/group/invite | 生成邀请码 | ```{ groupId: int, expireTime: int }``` | ```{ code: int, message: string, data: { inviteCode: string, expireAt: string } }``` | 群组管理员 |
| POST | /api/oss/group/member/remove | 移除成员 | ```{ groupId: int, userId: int }``` | ```{ code: int, message: string }``` | 群组管理员 |
| GET | /api/oss/group/stats?groupId=1 | 获取群组统计 | 无 | ```{ code: int, message: string, data: { projectCount: int, fileCount: int, totalSize: int } }``` | 群组成员 |

#### 项目管理接口

| 方法 | 路径 | 描述 | 请求参数 | 响应 | 权限要求 |
|------|------|------|---------|------|----------|
| POST | /api/oss/project/create | 创建项目 | ```{ groupId: int, name: string, description: string }``` | ```{ code: int, message: string, data: { projectId: int, pathPrefix: string } }``` | 群组管理员 |
| GET | /api/oss/project/list?groupId=1&page=1&size=10 | 获取项目列表 | 无 | ```{ code: int, message: string, data: { projects: array, total: int } }``` | 群组成员 |
| POST | /api/oss/project/permission | 设置项目权限 | ```{ projectId: int, userId: int, role: string }``` | ```{ code: int, message: string }``` | 项目管理员 |
| GET | /api/oss/project/users?projectId=1&page=1&size=10 | 获取项目用户 | 无 | ```{ code: int, message: string, data: { users: array, total: int } }``` | 项目成员 |
| POST | /api/oss/project/update | 更新项目信息 | ```{ projectId: int, name: string, description: string }``` | ```{ code: int, message: string }``` | 项目管理员 |
| POST | /api/oss/project/delete | 删除项目(逻辑删除) | ```{ projectId: int }``` | ```{ code: int, message: string }``` | 项目管理员 |
| GET | /api/oss/project/stats?projectId=1 | 项目存储统计 | 无 | ```{ code: int, message: string, data: { fileCount: int, folderCount: int, totalSize: int } }``` | 项目成员 |

#### 文件操作接口

| 方法 | 路径 | 描述 | 请求参数 | 响应 | 权限要求 |
|------|------|------|---------|------|----------|
| POST | /api/oss/file/instant-check | 秒传检查 | ```{ projectId: int, fileHash: string, fileName: string, filePath: string }``` | ```{ code: int, message: string, data: { canInstant: bool, fileId: int } }``` | 项目上传者 |
| POST | /api/oss/file/upload-token | 获取上传Token | ```{ projectId: int, fileName: string, filePath: string, fileSize: int, fileHash: string }``` | ```{ code: int, message: string, data: { uploadUrl: string, token: string, partSize: int, partCount: int } }``` | 项目上传者 |
| POST | /api/oss/file/upload-confirm | 确认上传 | ```{ token: string, parts: [{ partNumber: int, etag: string }] }``` | ```{ code: int, message: string, data: { fileId: int } }``` | 项目上传者 |
| GET | /api/oss/file/list?projectId=1&path=/folder&page=1&size=10&orderBy=name&order=asc | 文件列表 | 无 | ```{ code: int, message: string, data: { files: array, folders: array, total: int } }``` | 项目成员 |
| POST | /api/oss/file/folder/create | 创建文件夹 | ```{ projectId: int, folderPath: string }``` | ```{ code: int, message: string, data: { folderId: int } }``` | 项目上传者 |
| GET | /api/oss/file/download-token?fileId=1 | 获取下载Token | 无 | ```{ code: int, message: string, data: { downloadUrl: string, fileName: string, expireAt: string } }``` | 项目成员 |
| POST | /api/oss/file/delete | 删除文件(逻辑删除) | ```{ fileId: int, permanent: bool }``` | ```{ code: int, message: string }``` | 文件所有者或项目管理员 |
| POST | /api/oss/file/restore | 恢复已删除文件 | ```{ fileId: int }``` | ```{ code: int, message: string }``` | 文件所有者或项目管理员 |
| GET | /api/oss/file/trash?projectId=1&page=1&size=10 | 回收站列表 | 无 | ```{ code: int, message: string, data: { files: array, total: int } }``` | 项目成员 |
| POST | /api/oss/file/move | 移动文件 | ```{ fileId: int, targetPath: string }``` | ```{ code: int, message: string }``` | 文件所有者或项目管理员 |
| POST | /api/oss/file/copy | 复制文件 | ```{ fileId: int, targetPath: string }``` | ```{ code: int, message: string, data: { newFileId: int } }``` | 项目上传者 |
| GET | /api/oss/file/versions?fileId=1&page=1&size=10 | 获取文件版本 | 无 | ```{ code: int, message: string, data: { versions: array, total: int } }``` | 项目成员 |
| POST | /api/oss/file/revert | 回滚到历史版本 | ```{ fileId: int, versionId: int }``` | ```{ code: int, message: string }``` | 文件所有者或项目管理员 |
| POST | /api/oss/file/share | 分享文件 | ```{ fileId: int, password: string, expireTime: int, downloadLimit: int }``` | ```{ code: int, message: string, data: { shareCode: string, shareUrl: string } }``` | 项目成员 |
| GET | /api/oss/file/share/list?page=1&size=10 | 获取自己的分享 | 无 | ```{ code: int, message: string, data: { shares: array, total: int } }``` | 已登录 |
| POST | /api/oss/file/share/cancel | 取消分享(逻辑删除) | ```{ shareId: int }``` | ```{ code: int, message: string }``` | 分享创建者 |
| GET | /api/oss/file/share/info?shareCode=abc123 | 获取分享信息 | 无 | ```{ code: int, message: string, data: { fileInfo: object, needPassword: bool } }``` | 无 |
| POST | /api/oss/file/share/access | 访问分享 | ```{ shareCode: string, password: string }``` | ```{ code: int, message: string, data: { downloadUrl: string } }``` | 无 |

#### 日志查询接口

| 方法 | 路径 | 描述 | 请求参数 | 响应 | 权限要求 |
|------|------|------|---------|------|----------|
| GET | /api/oss/log/list?projectId=1&userId=2&operation=upload&startTime=2023-01-01&endTime=2023-01-31&page=1&size=10 | 日志列表 | 无 | ```{ code: int, message: string, data: { logs: array, total: int } }``` | 项目管理员 |
| GET | /api/oss/log/export?projectId=1&userId=2&operation=upload&startTime=2023-01-01&endTime=2023-01-31&format=csv | 导出日志 | 无 | 文件流 | 项目管理员 |
| GET | /api/oss/log/stats?projectId=1&timeRange=month | 操作统计 | 无 | ```{ code: int, message: string, data: { uploadCount: int, downloadCount: int, deleteCount: int, byUser: array } }``` | 项目管理员 |

### 3. 文件上传流程接口详解

#### 秒传机制
1. 客户端计算文件哈希值
2. 调用`POST /api/oss/file/instant-check`接口验证该哈希值是否存在
3. 如存在，服务端直接创建文件引用记录，无需实际上传
4. 响应中返回`canInstant: true`和文件ID

#### 标准上传流程
1. 客户端调用`POST /api/oss/file/upload-token`获取上传凭证
2. 服务端生成预签名URL和令牌，设置有效期（通常为15分钟）
3. 客户端使用预签名URL直接上传文件到MinIO
4. 上传完成后，客户端调用`POST /api/oss/file/upload-confirm`确认完成
5. 服务端验证上传状态，创建文件元数据记录

#### 分片上传流程
1. 客户端请求上传Token时，服务端根据文件大小决定是否需要分片
2. 对于大文件（如>10MB），返回分片大小和分片数量
3. 客户端按分片大小拆分文件，分别上传
4. 分片上传完成后，提交所有分片信息（ETag和分片号）
5. 服务端合并分片，创建文件记录

## 五、安全设计

### 1. 身份认证
- JWT令牌认证
- 密码加密存储
- 登录风控与防暴力破解

### 2. 传输安全
- HTTPS加密传输
- 预签名URL访问控制
- 文件完整性校验

### 3. 数据安全
- 服务端加密存储
- 数据备份策略
- 敏感信息保护

### 4. 操作安全
- 权限精细控制
- 操作审计日志
- 敏感操作二次验证
- 所有删除操作均为逻辑删除，支持恢复

## 六、性能优化

### 1. 数据库优化
- 索引优化
- 读写分离
- 连接池管理

### 2. 缓存策略
- 多级缓存设计
- 热点数据缓存
- 缓存失效策略

### 3. 并发处理
- Goroutine并发处理
- 连接复用
- 任务队列管理

### 4. 文件传输优化
- 分片上传
- 断点续传
- CDN加速支持

## 七、扩展设计

### 1. 服务扩展性
- 无状态服务设计
- 服务发现与负载均衡
- 配置中心集成

### 2. 存储扩展性
- MinIO分布式部署
- 多存储后端支持
- 分层存储策略

### 3. 功能扩展性
- 插件化架构
- 事件驱动设计
- 第三方集成接口

## 八、监控与运维

### 1. 系统监控
- 服务健康检查
- 资源使用监控
- 性能指标收集

### 2. 业务监控
- API调用统计
- 错误率监控
- 用户行为分析

### 3. 告警机制
- 多级告警策略
- 多渠道通知
- 自动恢复检测

### 4. 日志管理
- 日志聚合
- 日志分析
- 日志归档 