# Claude-ImgBed

一个使用 Claude Code 开发的轻量级、高性能图床系统，采用 Golang + React 构建，支持单二进制部署。

> 🤖 本项目完全由 [Claude Code](https://claude.ai/code) 辅助开发完成

## ✨ 功能特性

### 核心功能
- ✅ **多种上传方式**: 点击选择、Ctrl+V 粘贴、拖拽上传
- ✅ **批量上传**: 同时上传多张图片
- ✅ **实时进度**: 上传进度实时显示
- ✅ **图片预览**: 上传前后预览
- ✅ **一键复制**: 快速复制图片直链
- ✅ **最近上传**: 显示最近上传的图片

### 图片处理
- ✅ **自动压缩**: 质量 90，保持清晰度
- ✅ **智能缩放**: 大于 2000px 自动等比缩放
- ✅ **缩略图**: 自动生成 300x300 缩略图
- ✅ **格式支持**: JPG、PNG、GIF、WebP

### 安全特性
- ✅ **Bearer Token 认证**: 防止未授权访问
- ✅ **频率限制**: 每分钟 10 次上传限制
- ✅ **文件验证**: 魔数检查 + MIME 类型验证
- ✅ **大小限制**: 5MB 硬限制
- ✅ **CORS 支持**: 跨域访问控制

### 性能优化
- ✅ **CDN 友好**: 长期缓存头（1年）
- ✅ **并发处理**: Goroutine 并发处理
- ✅ **内存缓存**: 最近上传记录缓存
- ✅ **单二进制**: 前端嵌入，一个文件部署

### 存储管理
- ✅ **按日期存储**: 年/月目录自动创建
- ✅ **文件命名**: YYYYMMDDhhmmss_uuid.ext
- ✅ **健康检查**: /health 端点

## 🚀 快速开始

### 1. 构建

```bash
cd .
make build
```

### 2. 配置

编辑 `config.yaml`：

```yaml
auth:
  token: "your-secret-token-here"  # ⚠️ 必须修改
```

### 3. 运行

```bash
./imgbed
```

### 4. 访问

打开浏览器：http://localhost:8080

详细步骤请查看 [QUICKSTART.md](QUICKSTART.md)

## 📖 文档

- [QUICKSTART.md](QUICKSTART.md) - 快速开始指南
- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构设计文档
- [DEPLOYMENT.md](DEPLOYMENT.md) - 详细部署指南
- [tests/README.md](tests/README.md) - 测试文档

## API 文档

### 上传图片

```bash
POST /api/upload
Authorization: Bearer your-secret-token-here
Content-Type: multipart/form-data

# 示例
curl -X POST http://localhost:8080/api/upload \
  -H "Authorization: Bearer your-secret-token-here" \
  -F "file=@image.jpg"
```

响应：
```json
{
  "success": true,
  "data": {
    "url": "http://localhost:8080/images/2026/03/20260302123456_abc12345.jpg",
    "thumbnail": "http://localhost:8080/images/2026/03/20260302123456_abc12345_thumb.jpg",
    "filename": "20260302123456_abc12345.jpg",
    "size": 1024000,
    "width": 1920,
    "height": 1080,
    "uploadedAt": "2026-03-02T12:34:56Z"
  }
}
```

### 获取最近上传

```bash
GET /api/recent
Authorization: Bearer your-secret-token-here

# 示例
curl http://localhost:8080/api/recent \
  -H "Authorization: Bearer your-secret-token-here"
```

### 访问图片

```bash
GET /images/:year/:month/:filename

# 示例
curl http://localhost:8080/images/2026/03/20260302123456_abc12345.jpg
```

### 健康检查

```bash
GET /health

# 示例
curl http://localhost:8080/health
```

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.port` | 服务端口 | 8080 |
| `server.host` | 监听地址 | 0.0.0.0 |
| `auth.token` | 认证令牌 | change-me-in-production |
| `upload.max_size` | 最大文件大小（字节） | 5242880 (5MB) |
| `upload.storage_path` | 存储路径 | ./uploads |
| `image.max_dimension` | 最大尺寸（像素） | 2000 |
| `image.quality` | 压缩质量 | 90 |
| `image.thumbnail_size` | 缩略图尺寸 | 300 |
| `rate_limit.requests_per_minute` | 每分钟请求限制 | 10 |
| `cache.recent_uploads_size` | 缓存记录数 | 100 |

## 环境变量

可以通过环境变量覆盖配置：

```bash
export AUTH_TOKEN="your-secret-token"
./imgbed
```

## Docker 部署

```bash
# 构建镜像
docker build -t imgbed .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/uploads:/app/uploads \
  -e AUTH_TOKEN="your-secret-token" \
  --name imgbed \
  imgbed
```

## 项目结构

```
imgbed/
├── cmd/server/main.go           # 程序入口
├── internal/
│   ├── api/
│   │   ├── handler.go           # API 处理器
│   │   ├── middleware.go        # 中间件
│   │   └── router.go            # 路由配置
│   ├── auth/
│   │   └── auth.go              # 认证
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── image/
│   │   ├── processor.go         # 图片处理
│   │   ├── storage.go           # 存储管理
│   │   └── validator.go         # 文件验证
│   ├── models/
│   │   └── upload.go            # 数据模型
│   └── ratelimit/
│       └── limiter.go           # 频率限制
├── config.yaml                  # 配置文件
├── go.mod                       # Go 模块
├── Makefile                     # 构建脚本
├── Dockerfile                   # Docker 配置
└── README.md                    # 文档
```

## 安全特性

1. **文件验证**：魔数检查 + MIME 类型验证
2. **大小限制**：5MB 硬限制
3. **频率限制**：防止滥用（每分钟10次）
4. **认证保护**：Bearer Token 认证
5. **CORS 配置**：跨域访问控制

## 性能优化

1. **并发处理**：使用 Goroutine 处理图片
2. **CDN 缓存**：设置长期缓存头（1年）
3. **图片压缩**：自动压缩和缩放
4. **内存缓存**：最近上传记录缓存

## 依赖包

- `github.com/gin-gonic/gin` - Web 框架
- `github.com/disintegration/imaging` - 图片处理
- `github.com/google/uuid` - UUID 生成
- `github.com/spf13/viper` - 配置管理
- `golang.org/x/time/rate` - 频率限制

## 开发

```bash
# 运行测试
make test

# 清理构建产物
make clean

# 下载依赖
make deps
```

## License

MIT
