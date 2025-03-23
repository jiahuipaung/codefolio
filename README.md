# Codefolio

Codefolio 是一个基于 Go 语言开发的个人作品集展示平台。

## 功能特性

- 用户认证（邮箱注册/登录）
- JWT 身份验证
- 简历上传和展示（开发中）

## 技术栈

- Go 1.21
- Gin Web 框架
- GORM ORM
- PostgreSQL 数据库
- JWT 认证

## 项目结构

```
.
├── cmd/                    # 应用程序入口
├── internal/              # 内部包
│   ├── domain/           # 领域模型
│   ├── repository/       # 数据访问层
│   ├── service/          # 业务逻辑层
│   └── handler/          # HTTP 处理器
├── pkg/                   # 公共包
├── config/               # 配置文件
└── migrations/           # 数据库迁移文件
```

## 开始使用

1. 克隆项目
```bash
git clone [repository-url]
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，填入必要的配置信息
```

4. 运行项目
```bash
go run cmd/main.go
```

## API 文档

### 用户认证

- POST /api/v1/auth/register - 用户注册
- POST /api/v1/auth/login - 用户登录
- GET /api/v1/auth/me - 获取当前用户信息

## 开发计划

- [x] 用户认证系统
- [ ] 简历上传功能
- [ ] 简历展示功能
- [ ] 作品集管理 