
#Go镜像作为构建环境
FROM golang:1.25-alpine AS builder
# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
# 只有go.mod和go.sum文件变化时才会重新下载依赖
RUN go mod download

# 复制源代码
COPY . .
# cgo都来了（害怕），关闭cgo确保静态编译
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/myapp/main.go 

# 使用一个更小的镜像来运行应用
FROM alpine:latest

# 设置工作目录
WORKDIR /app
# 从builder阶段复制可执行二进制文件
COPY --from=builder /app/server .
# 暴露应用运行端口
EXPOSE 8080         
# 设置容器启动命令
ENTRYPOINT ["/app/server"]