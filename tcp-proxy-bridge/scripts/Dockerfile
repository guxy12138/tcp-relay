# Dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用（禁用CGO，静态链接）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装必要的工具（健康检查需要wget）
RUN apk --no-cache add ca-certificates wget

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
# 复制配置文件
COPY --from=builder /app/configs configs/

# 暴露端口
EXPOSE 8080 9999

# 启动命令
CMD ["./main"]