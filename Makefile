.PHONY: build run dev clean docker-build docker-compose init-db create-db swagger docs api-docs all-docs

# 默认目标
all: build

# 构建应用
build:
	go build -o bin/oss-backend main.go

# 运行应用
run: build
	./bin/oss-backend

# 开发模式运行
dev:
	go run main.go

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
	@swag init -g main.go -o docs/swagger
else
	@go install github.com/swaggo/swag/cmd/swag@latest
	@$(shell go env GOPATH)/bin/swag init -g main.go -o docs/swagger
endif

# 生成API文档（基于Swagger）
api-docs: swagger
	@echo "生成API文档完成。可在运行应用后访问: http://localhost:8080/swagger/index.html"

# 生成代码文档
code-docs:
	@echo "生成Go代码文档..."
ifeq ($(OS),Windows_NT)
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "文档服务已启动，请访问: http://localhost:6060/pkg/oss-backend/"
	@godoc -http=:6060
else
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "文档服务已启动，请访问: http://localhost:6060/pkg/oss-backend/"
	@$(shell go env GOPATH)/bin/godoc -http=:6060
endif

# 生成所有文档
all-docs: swagger api-docs
	@echo "所有文档已生成"

# 启动文档服务器
serve-docs: all-docs
	@echo "启动文档服务器..."
ifeq ($(OS),Windows_NT)
	@cd docs/swagger && python -m http.server 8090
else
	@cd docs/swagger && python3 -m http.server 8090
endif
	@echo "文档服务器已启动，请访问: http://localhost:8090"

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
	@echo "  make api-docs       - 生成API文档（包含Swagger）"
	@echo "  make code-docs      - 生成代码文档并启动godoc服务器"
	@echo "  make all-docs       - 生成所有文档"
	@echo "  make serve-docs     - 启动文档Web服务器" 