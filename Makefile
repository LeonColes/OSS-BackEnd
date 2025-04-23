.PHONY: build run dev clean docker-build docker-compose init-db create-db swagger

# 默认目标
all: build

# 构建应用
build:
	go build -o bin/oss-backend cmd/main.go

# 运行应用
run: build
	./bin/oss-backend

# 开发模式运行
dev:
	go run cmd/main.go

# 清理构建文件
clean:
	rm -rf bin/

# 启动开发环境的数据库和服务
dev-env-up:
	docker-compose up -d mysql redis minio

# 停止开发环境
dev-env-down:
	docker-compose down

# 重置开发环境
dev-env-reset: dev-env-down
	docker-compose up -d --force-recreate mysql redis minio

# 构建Docker镜像
docker-build:
	docker build -t oss-backend:latest .

# 运行Docker Compose环境
docker-compose:
	docker-compose up -d

# 停止Docker Compose环境
docker-compose-down:
	docker-compose down

# 显示日志
logs:
	docker-compose logs -f

# 生成Swagger文档
swagger:
	@echo "生成Swagger文档..."
ifeq ($(OS),Windows_NT)
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/main.go -o docs/swagger
else
	@go install github.com/swaggo/swag/cmd/swag@latest
	@$(shell go env GOPATH)/bin/swag init -g cmd/main.go -o docs/swagger
endif

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make build          - 构建应用"
	@echo "  make run            - 构建并运行应用"
	@echo "  make dev            - 开发模式运行应用"
	@echo "  make clean          - 清理构建文件"
	@echo "  make dev-env-up     - 启动开发环境的数据库和服务"
	@echo "  make dev-env-down   - 停止开发环境"
	@echo "  make dev-env-reset  - 重置开发环境"
	@echo "  make docker-build   - 构建Docker镜像"
	@echo "  make docker-compose - 运行Docker Compose环境"
	@echo "  make docker-compose-down - 停止Docker Compose环境"
	@echo "  make logs           - 显示Docker Compose日志"
	@echo "  make init-db        - 通过Docker初始化数据库"
	@echo "  make create-db      - 不依赖Docker创建和初始化数据库"
	@echo "  make swagger        - 生成Swagger文档"