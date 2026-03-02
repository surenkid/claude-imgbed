# 快速开始指南

## 一键构建和运行

### 1. 构建项目

```bash
# 进入项目目录
cd /config/Projects/imgbed

# 一键构建（前端+后端）
make build
```

这会自动：
- 安装前端依赖（npm install）
- 构建前端（生成 web/dist/）
- 下载 Go 依赖
- 编译后端并嵌入前端（生成 imgbed 二进制文件）

### 2. 配置认证密钥

编辑 `config.yaml`，修改认证 token：

```yaml
auth:
  token: "your-secret-token-here"  # 改成你自己的密钥
```

### 3. 运行

```bash
./imgbed
```

服务将在 http://localhost:8080 启动。

### 4. 访问

打开浏览器访问：http://localhost:8080

使用你在 config.yaml 中设置的 token 登录。

## 开发模式

### 后端开发

```bash
make run
# 或
go run cmd/server/main.go
```

### 前端开发（带热重载）

```bash
cd web
npm install
npm run dev
```

前端开发服务器会在 http://localhost:5173 启动，API 请求会自动代理到后端。

## 生产部署

### 优化构建

```bash
make build-prod
```

这会生成优化后的二进制文件（更小的体积）。

### 多平台构建

```bash
make build-all
```

生成 Linux、macOS、Windows 三个平台的可执行文件。

## 测试

```bash
make test
```

## 清理

```bash
make clean
```

## 完整文档

- [README.md](README.md) - 项目介绍和 API 文档
- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构设计文档
- [DEPLOYMENT.md](DEPLOYMENT.md) - 详细部署指南
- [tests/README.md](tests/README.md) - 测试文档

## 常见问题

**Q: 构建失败怎么办？**
- 确保已安装 Go 1.21+ 和 Node.js 18+
- 运行 `make deps` 下载依赖
- 检查网络连接

**Q: 如何修改端口？**
- 编辑 `config.yaml` 中的 `server.port`

**Q: 如何部署到服务器？**
- 参考 [DEPLOYMENT.md](DEPLOYMENT.md) 的详细部署指南

**Q: 支持 Docker 吗？**
- 是的，运行 `docker build -t imgbed .` 构建镜像
