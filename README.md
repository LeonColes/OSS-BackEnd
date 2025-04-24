# OSS-Backend

基于Golang的对象存储服务后端，提供文件上传、下载、管理等功能。

## 技术栈

- **语言**: Golang 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+
- **对象存储**: MinIO
- **API文档**: Swagger
- **日志**: Zap
- **配置管理**: Viper

## 系统架构

- **API层**: 处理HTTP请求，参数校验，权限检查
- **Service层**: 实现业务逻辑
- **Repository层**: 数据访问抽象
- **Model层**: 数据模型定义
- **Util层**: 工具函数

## 开发环境设置

### 前置要求

- Go 1.21+
- Docker & Docker Compose

### 快速启动

1. 克隆代码库:

```bash
git clone https://github.com/leoncoles/oss-backend.git
cd oss-backend
```

2. 启动依赖服务(MySQL, Redis, MinIO):

使用Docker Compose:

```bash
docker-compose up -d
```

3. 编译并运行应用:

**Linux/Mac环境:**

如果安装了Make:
```bash
make dev
```

没有安装Make:
```bash
go run cmd/server/main.go
```

**Windows环境:**
```bash
go run cmd/server/main.go
```

4. 访问API:

API服务默认运行在 `http://localhost:8080`

### 环境配置

- 配置文件位于 `configs/` 目录
- 开发环境使用 `config.dev.yaml`
- 生产环境使用 `config.yaml`
- 环境变量可覆盖配置文件中的设置

### 开发命令

项目提供了一系列便捷的Make命令(仅Linux/Mac环境):

- `make build` - 构建应用
- `make run` - 构建并运行应用
- `make dev` - 开发模式运行应用
- `make clean` - 清理构建文件
- `make dev-env-up` - 启动开发环境依赖服务
- `make dev-env-down` - 停止开发环境
- `make dev-env-reset` - 重置开发环境
- `make help` - 显示所有可用命令

**Windows环境替代命令:**
- 构建应用: `go build -o bin/oss-backend.exe cmd/server/main.go`
- 运行应用: `go run cmd/server/main.go`
- 启动依赖: `docker-compose up -d mysql redis minio`
- 停止依赖: `docker-compose down`

## 生产环境部署

### 使用Docker Compose

1. 构建项目镜像:

```bash
docker build -t oss-backend:latest .
```

2. 启动所有服务:

```bash
docker-compose up -d
```

3. 查看日志:

```bash
docker-compose logs -f oss-backend
```

### 使用独立部署

1. 构建可执行文件:

**Linux/Mac:**
```bash
go build -o bin/oss-backend cmd/server/main.go
```

**Windows:**
```bash
go build -o bin\oss-backend.exe cmd\server\main.go
```

2. 将二进制文件和配置文件复制到服务器:

```bash
scp bin/oss-backend user@server:/path/to/app/
scp configs/config.yaml user@server:/path/to/app/configs/
```

3. 设置环境变量或修改配置文件以适应生产环境

4. 运行应用:

```bash
cd /path/to/app && ./oss-backend
```

## 项目文档

项目提供了全面的文档，位于`docs/`目录下：

1. **[架构设计文档](docs/architecture.md)**
   - 系统概述
   - 架构设计原则
   - 整体架构
   - 技术栈选型
   - 核心模块设计
   - 存储设计
   - 认证与授权设计（含角色与权限管理）
   - 部署架构
   - 性能与扩展性
   - 安全设计

2. **[API文档](docs/api.md)**
   - 接口规范
   - 角色权限列表
   - 核心接口说明
   - Swagger使用指南

3. **[贡献指南](docs/contributing.md)**
   - 行为准则
   - 开发流程
   - 代码风格
   - 提交规范
   - Pull Request流程
   - 项目结构
   - 开发环境设置

### API文档

#### Swagger接口文档

开发环境下，可通过浏览器访问Swagger UI:

```
http://localhost:8080/swagger/index.html
```

#### Apifox导入指南

可以轻松导入Swagger文档到Apifox进行API测试：

1. 确保OSS-Backend服务正在运行
2. 打开Apifox
3. 点击 "导入/新建" > "导入数据" > "OpenAPI (Swagger)"
4. 选择导入方式：
   - 通过URL: `http://localhost:8080/swagger/doc.json`
   - 或通过文件: `docs/swagger/swagger.json`
5. 完成导入设置，点击"确定"