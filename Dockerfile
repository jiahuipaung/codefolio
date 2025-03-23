# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

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