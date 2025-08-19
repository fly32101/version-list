# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-version .

# 最终阶段
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates git

# 设置工作目录
WORKDIR /root/

# 从 builder 阶段复制二进制文件
COPY --from=builder /app/go-version .

# 创建 Go 版本安装目录
RUN mkdir -p /go-versions

# 设置环境变量
ENV GO_VERSIONS_PATH=/go-versions

# 暴露端口（如果需要）
# EXPOSE 8080

# 运行应用
CMD ["./go-version"]
