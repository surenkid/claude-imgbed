# 图床系统部署指南

## 构建步骤

### 1. 前置要求
- Go 1.21+
- Node.js 18+
- npm 或 yarn

### 2. 完整构建流程

```bash
# 克隆或进入项目目录
# 进入项目目录
cd claude-imgbed

# 方式一：使用 Makefile 一键构建（推荐）
make build

# 方式二：手动分步构建
# 步骤1: 构建前端
cd web
npm install
npm run build
cd ..

# 步骤2: 构建后端（会自动嵌入前端）
go mod download
go build -o imgbed cmd/server/main.go
```

### 3. 生产环境优化构建

```bash
# 优化构建（减小二进制文件大小）
make build-prod

# 或手动执行
CGO_ENABLED=0 go build -ldflags="-s -w" -o imgbed cmd/server/main.go
```

### 4. 多平台构建

```bash
# 构建 Linux/macOS/Windows 版本
make build-all

# 或手动执行
GOOS=linux GOARCH=amd64 go build -o imgbed-linux-amd64 cmd/server/main.go
GOOS=darwin GOARCH=amd64 go build -o imgbed-darwin-amd64 cmd/server/main.go
GOOS=windows GOARCH=amd64 go build -o imgbed-windows-amd64.exe cmd/server/main.go
```

## 配置

### 1. 修改配置文件

编辑 `config.yaml`：

```yaml
server:
  port: "8080"
  host: "0.0.0.0"

auth:
  token: "your-secret-token-here"  # ⚠️ 必须修改！

upload:
  max_size: 5242880  # 5MB
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
  storage_path: "./uploads"

image:
  max_dimension: 2000
  quality: 90
  thumbnail_size: 300

rate_limit:
  requests_per_minute: 10
  burst: 5

cache:
  recent_uploads_size: 100
```

### 2. 环境变量（可选）

```bash
export AUTH_TOKEN="your-secret-token"
export SERVER_PORT="8080"
```

## 运行

### 开发模式

```bash
# 后端开发
make run
# 或
go run cmd/server/main.go

# 前端开发（带热重载）
cd web
npm run dev
```

### 生产模式

```bash
# 直接运行二进制文件
./imgbed

# 后台运行
nohup ./imgbed > imgbed.log 2>&1 &

# 使用 systemd（推荐）
sudo systemctl start imgbed
```

## Docker 部署

### 1. 构建镜像

```bash
docker build -t imgbed:latest .
```

### 2. 运行容器

```bash
docker run -d \
  --name imgbed \
  -p 8080:8080 \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -e AUTH_TOKEN="your-secret-token" \
  imgbed:latest
```

### 3. Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  imgbed:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./uploads:/app/uploads
      - ./config.yaml:/app/config.yaml
    environment:
      - AUTH_TOKEN=your-secret-token
    restart: unless-stopped
```

运行：
```bash
docker-compose up -d
```

## Systemd 服务配置

创建 `/etc/systemd/system/imgbed.service`：

```ini
[Unit]
Description=Image Hosting Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/imgbed
ExecStart=/opt/imgbed/imgbed
Restart=on-failure
RestartSec=5s

Environment="AUTH_TOKEN=your-secret-token"

[Install]
WantedBy=multi-user.target
```

启用服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable imgbed
sudo systemctl start imgbed
sudo systemctl status imgbed
```

## Nginx 反向代理

```nginx
server {
    listen 80;
    server_name img.yourdomain.com;

    client_max_body_size 10M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 图片缓存
    location /images/ {
        proxy_pass http://localhost:8080;
        proxy_cache_valid 200 1y;
        add_header Cache-Control "public, max-age=31536000";
    }
}
```

## Cloudflare CDN 配置

### 1. DNS 设置
- 添加 A 记录指向服务器 IP
- 启用 Cloudflare 代理（橙色云朵）

### 2. 缓存规则
在 Cloudflare Dashboard → Rules → Page Rules：

```
URL: img.yourdomain.com/images/*
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 year
  - Browser Cache TTL: 1 year
```

### 3. 性能优化
- 启用 Auto Minify (JS, CSS, HTML)
- 启用 Brotli 压缩
- 启用 HTTP/2
- 启用 Polish (图片优化)
- 启用 Mirage (移动端优化)

### 4. 安全设置
- 启用 HTTPS
- 设置 SSL/TLS 为 Full (strict)
- 启用 Always Use HTTPS
- 配置 WAF 规则

## 测试

```bash
# 运行所有测试
make test

# 或手动执行
go test ./tests/... -v

# 运行性能测试
go test ./tests/... -bench=. -benchmem
```

## 监控和日志

### 健康检查

```bash
curl http://localhost:8080/health
```

### 日志查看

```bash
# 查看实时日志
tail -f imgbed.log

# systemd 日志
sudo journalctl -u imgbed -f
```

## 备份

### 备份上传的图片

```bash
# 定期备份 uploads 目录
tar -czf uploads-backup-$(date +%Y%m%d).tar.gz uploads/

# 使用 rsync 同步到远程服务器
rsync -avz uploads/ user@backup-server:/backup/imgbed/
```

### 自动备份脚本

创建 `/opt/imgbed/backup.sh`：

```bash
#!/bin/bash
BACKUP_DIR="/backup/imgbed"
DATE=$(date +%Y%m%d-%H%M%S)

# 创建备份
tar -czf $BACKUP_DIR/uploads-$DATE.tar.gz /opt/imgbed/uploads/

# 保留最近30天的备份
find $BACKUP_DIR -name "uploads-*.tar.gz" -mtime +30 -delete
```

添加到 crontab：
```bash
0 2 * * * /opt/imgbed/backup.sh
```

## 故障排查

### 1. 端口被占用
```bash
# 查看端口占用
sudo lsof -i :8080

# 修改配置文件中的端口
```

### 2. 权限问题
```bash
# 确保 uploads 目录可写
chmod 755 uploads/
chown -R www-data:www-data uploads/
```

### 3. 内存不足
```bash
# 查看内存使用
free -h

# 限制 Go 内存使用
GOGC=50 ./imgbed
```

### 4. 磁盘空间不足
```bash
# 查看磁盘使用
df -h

# 清理旧图片（可选）
find uploads/ -type f -mtime +365 -delete
```

## 性能调优

### 1. Go 运行时优化

```bash
# 设置 GOMAXPROCS（默认为 CPU 核心数）
GOMAXPROCS=4 ./imgbed

# 调整 GC 频率
GOGC=100 ./imgbed
```

### 2. 系统优化

```bash
# 增加文件描述符限制
ulimit -n 65535

# 调整 TCP 参数
sudo sysctl -w net.core.somaxconn=1024
```

## 安全加固

### 1. 防火墙配置

```bash
# 只允许必要端口
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. 定期更新

```bash
# 更新依赖
go get -u ./...
go mod tidy

# 重新构建
make build
```

### 3. 限制上传来源

修改 `internal/api/middleware.go` 添加 IP 白名单或 Referer 检查。

## 升级指南

```bash
# 1. 备份当前版本
cp imgbed imgbed.backup
tar -czf uploads-backup.tar.gz uploads/

# 2. 拉取新代码
git pull

# 3. 重新构建
make build

# 4. 重启服务
sudo systemctl restart imgbed

# 5. 验证
curl http://localhost:8080/health
```

## 常见问题

**Q: 如何修改最大上传大小？**
A: 修改 `config.yaml` 中的 `upload.max_size`，单位为字节。

**Q: 如何增加上传频率限制？**
A: 修改 `config.yaml` 中的 `rate_limit.requests_per_minute`。

**Q: 如何清理旧图片？**
A: 可以编写定时任务删除超过一定时间的图片，或使用对象存储的生命周期策略。

**Q: 支持对象存储吗？**
A: 当前版本使用本地存储，可以扩展支持 S3/OSS 等对象存储。

**Q: 如何启用 HTTPS？**
A: 使用 Nginx 反向代理配置 SSL 证书，或使用 Cloudflare 的 SSL。

## 联系支持

如有问题，请查看：
- 项目文档：README.md
- 架构文档：ARCHITECTURE.md
- 测试文档：tests/README.md
