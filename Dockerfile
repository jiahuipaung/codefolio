# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 设置 GOPROXY 和 GOSUMDB
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GO111MODULE=on

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制 go.mod
COPY go.mod ./

# 下载依赖并生成 go.sum
RUN go mod tidy && \
    go mod download && \
    go mod verify

# 复制源代码
COPY . .

# 验证依赖
RUN go mod verify

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"] 