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

## API文档

API文档使用Swagger生成，可通过以下方式访问:

```
http://localhost:8080/swagger/index.html
```

### Swagger目录文件说明

Swagger文档在 `docs/swagger/` 目录下，包含以下重要文件：

- `swagger.json` - OpenAPI规范的JSON格式文档，用于工具导入
- `swagger.yaml` - OpenAPI规范的YAML格式文档，与JSON内容一致但格式更易读
- `docs.go` - Go代码形式的Swagger文档，由swag工具自动生成

生成或更新Swagger文档：

```bash
# Linux/Mac
make swagger

# Windows
swag init -g cmd/server/main.go -o docs/swagger
```

### Apifox导入指南

[Apifox](https://www.apifox.cn/) 是一个优秀的API设计、开发、测试一体化工具，可以轻松导入Swagger文档。

#### 前提条件

1. 已安装 [Apifox](https://www.apifox.cn/download/)
2. OSS-Backend服务已成功运行

#### 方法一：通过Swagger JSON文件导入

1. 确保已经生成了Swagger文档:
   ```bash
   make swagger
   ```

2. 导入步骤：
   - 打开Apifox
   - 点击 "导入/新建" > "导入数据" > "OpenAPI (Swagger)"
   - 选择 `docs/swagger/swagger.json` 文件导入
   - 完成导入设置，点击"确定"

#### 方法二：通过Swagger URL导入

1. 确保OSS-Backend服务正在运行:
   ```bash
   make dev
   ```

2. 导入步骤：
   - 打开Apifox
   - 点击 "导入/新建" > "导入数据" > "OpenAPI (Swagger)"
   - 选择"从URL导入"，填入：`http://localhost:8080/swagger-json/swagger.json`
   - 完成导入设置，点击"确定"

#### 方法三：解决导入问题

如果导入URL时遇到 `404 Not Found` 错误，请尝试以下步骤：

1. 确保服务器正在运行：
   ```bash
   make dev
   ```

2. 在浏览器中访问 `http://localhost:8080/swagger-json/swagger.json` 确认文件是否可访问

3. 如果文件不可访问，重新生成Swagger文档：
   ```bash
   make swagger
   # 或者在Windows系统上：
   swag init -g cmd/server/main.go -o docs/swagger
   ```

4. 重启服务器后再次尝试导入

5. 如果仍有问题，可以下载swagger.json文件，然后在Apifox中选择"导入本地文件"

#### 系统默认账户

导入完成后，你可以使用以下账户进行API测试：

- 管理员账户:
  - 邮箱: admin@example.com
  - 密码: admin123

#### API认证说明

在调用需要认证的API时，请先执行登录接口，然后将获取到的`access_token`添加到其他API的请求头中：

```
Authorization: Bearer {access_token}
```

在Apifox中，你可以设置全局请求头或环境变量来自动添加这个认证信息。

#### 多环境配置

Swagger文档定义了多个环境，可在Apifox中轻松切换：

1. **本地开发环境**：http://localhost:8080/api/oss
2. **测试环境**：http://test-api.example.com/api/oss
3. **生产环境**：http://api.example.com/api/oss

导入Apifox后，可以根据需要修改这些环境配置的URL。

## 许可证

[MIT License](LICENSE) 