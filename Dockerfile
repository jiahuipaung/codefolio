# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 设置 GOPROXY 和 GOSUMDB
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GO111MODULE=on

# 安装必要的构建工具
RUN apk add --no-cache git

# 设置工作目录
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

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 安装PDF转图片工具（poppler-utils包含pdftoppm）
RUN apk --no-cache add poppler-utils imagemagick ghostscript

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建默认的 .env 文件
RUN echo "# 服务器配置\nPORT=8080\nENV=development\n\n# 数据库配置\nDB_HOST=localhost\nDB_PORT=5432\nDB_USER=postgres\nDB_PASSWORD=postgres\nDB_NAME=codefolio\n\n# JWT配置\nJWT_SECRET=your-secret-key\nJWT_EXPIRATION=24h\n\n# 邮件配置\nSMTP_HOST=smtp.example.com\nSMTP_PORT=587\nSMTP_USER=your-email@example.com\nSMTP_PASSWORD=your-email-password" > .env

# 设置权限
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"] 