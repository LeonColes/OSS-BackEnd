# 使用golang官方镜像作为构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖和CI工具
RUN apk add --no-cache git gcc musl-dev

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o oss-backend ./cmd/server

# 使用Alpine作为运行阶段
FROM alpine:latest

# 添加运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区为亚洲/上海
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的应用
COPY --from=builder /app/oss-backend .

# 复制配置文件
COPY ./configs/config.yaml /app/configs/

# 创建必要的目录
RUN mkdir -p /app/uploads /app/temp

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./oss-backend"] 