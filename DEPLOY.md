# 部署指南

本指南将帮助您将 Codefolio 项目部署到服务器上。

## 服务器要求

- Linux 服务器（推荐 Ubuntu 20.04 或更高版本）
- Docker 和 Docker Compose
- 至少 2GB RAM
- 至少 20GB 存储空间

## 安装 Docker 和 Docker Compose

### Ubuntu

```bash
# 更新包列表
sudo apt update

# 安装必要的依赖
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common

# 添加 Docker 官方 GPG 密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# 添加 Docker 仓库
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 更新包列表
sudo apt update

# 安装 Docker
sudo apt install -y docker-ce docker-ce-cli containerd.io

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

## 部署步骤

1. 克隆项目到服务器：
```bash
git clone [repository-url]
cd codefolio
```

2. 配置环境变量：
```bash
cp .env.example .env
# 编辑 .env 文件，设置必要的环境变量
# 特别注意修改 JWT_SECRET 为安全的随机字符串
```

3. 构建和启动服务：
```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

4. 检查服务状态：
```bash
docker-compose ps
```

5. 查看应用日志：
```bash
docker-compose logs -f app
```

## 维护命令

### 停止服务
```bash
docker-compose down
```

### 重启服务
```bash
docker-compose restart
```

### 更新应用
```bash
# 拉取最新代码
git pull

# 重新构建和启动服务
docker-compose up -d --build
```

### 备份数据库
```bash
# 创建备份
docker-compose exec db pg_dump -U postgres codefolio > backup.sql

# 恢复备份
cat backup.sql | docker-compose exec -T db psql -U postgres codefolio
```

## 安全建议

1. 修改默认的数据库密码
2. 使用强密码作为 JWT_SECRET
3. 配置防火墙，只开放必要的端口
4. 定期更新系统和依赖包
5. 配置 SSL 证书（推荐使用 Let's Encrypt）

## 故障排除

1. 如果应用无法启动，检查日志：
```bash
docker-compose logs app
```

2. 如果数据库连接失败，检查数据库日志：
```bash
docker-compose logs db
```

3. 如果需要进入容器调试：
```bash
docker-compose exec app sh
```

## 监控

建议配置以下监控：

1. 服务器资源监控（CPU、内存、磁盘使用率）
2. 应用日志监控
3. 数据库性能监控
4. 应用健康检查

可以使用 Prometheus + Grafana 或 ELK Stack 等工具进行监控。 