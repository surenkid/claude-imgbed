# 图床系统架构设计

## 技术栈

### 后端
- **语言**: Golang 1.21+
- **Web框架**: Gin (轻量高性能)
- **图片处理**: imaging (纯Go实现，易于交叉编译)
- **限流**: golang.org/x/time/rate
- **配置**: viper

### 前端
- **框架**: React 18
- **样式**: TailwindCSS
- **构建**: Vite
- **状态管理**: React Hooks
- **HTTP客户端**: axios

### 打包
- **嵌入**: Go embed包
- **构建**: Makefile自动化

## 项目结构

```
imgbed/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── api/
│   │   ├── handler.go           # API处理器
│   │   ├── middleware.go        # 中间件（认证、限流、CORS）
│   │   └── router.go            # 路由配置
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── image/
│   │   ├── processor.go         # 图片压缩、缩放、缩略图
│   │   ├── storage.go           # 文件存储管理
│   │   └── validator.go         # 文件验证（格式、大小、魔数）
│   ├── auth/
│   │   └── auth.go              # 简单认证
│   ├── ratelimit/
│   │   └── limiter.go           # 频率限制
│   └── models/
│       └── upload.go            # 上传记录模型
├── web/                         # 前端项目
│   ├── src/
│   │   ├── components/
│   │   │   ├── Upload.jsx       # 上传组件
│   │   │   ├── ImagePreview.jsx # 图片预览
│   │   │   ├── RecentUploads.jsx# 最近上传
│   │   │   └── Auth.jsx         # 认证组件
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── public/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
├── static/                      # 前端构建产物（嵌入到二进制）
├── uploads/                     # 图片存储目录
│   └── YYYY/MM/                 # 按年月分文件夹
├── tests/
│   ├── api_test.go              # API测试
│   ├── image_test.go            # 图片处理测试
│   └── e2e_test.go              # E2E测试
├── config.yaml                  # 配置文件
├── go.mod
├── go.sum
├── Makefile                     # 构建脚本
├── Dockerfile                   # Docker配置（可选）
└── README.md
```

## API设计

### 认证
- **方式**: Bearer Token (固定密钥)
- **Header**: `Authorization: Bearer <token>`

### 端点

#### 1. 上传图片
```
POST /api/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

Body:
- file: 图片文件

Response:
{
  "success": true,
  "data": {
    "url": "https://yourdomain.com/images/2026/03/20260302abc123.jpg",
    "thumbnail": "https://yourdomain.com/images/2026/03/20260302abc123_thumb.jpg",
    "filename": "20260302abc123.jpg",
    "size": 1024000,
    "width": 1920,
    "height": 1080,
    "uploadedAt": "2026-03-02T12:00:00Z"
  }
}
```

#### 2. 获取最近上传
```
GET /api/recent
Authorization: Bearer <token>

Response:
{
  "success": true,
  "data": [
    {
      "url": "...",
      "thumbnail": "...",
      "uploadedAt": "..."
    }
  ]
}
```

#### 3. 访问图片
```
GET /images/:year/:month/:filename

Response: 图片文件
Headers:
- Cache-Control: public, max-age=31536000
- CDN-Cache-Control: max-age=31536000
```

#### 4. 健康检查
```
GET /health

Response:
{
  "status": "ok",
  "timestamp": "2026-03-02T12:00:00Z"
}
```

## 核心功能实现

### 1. 图片处理流程
```
上传 → 验证(格式/大小/魔数) → 压缩(质量90) → 缩放(>2000px) → 生成缩略图 → 存储 → 返回URL
```

### 2. 文件验证
- **格式**: jpg, png, gif, webp
- **大小**: 最大5MB
- **魔数检查**: 防止伪造扩展名
- **MIME类型**: 验证Content-Type

### 3. 图片压缩
- **质量**: 90
- **缩放**: 长边>2000px时等比例缩放至2000px
- **缩略图**: 300x300 (保持比例)

### 4. 存储策略
- **路径**: `uploads/YYYY/MM/YYYYMMDDhhmmss_uuid.ext`
- **命名**: 时间戳 + UUID前8位
- **自动创建**: 年月目录自动创建

### 5. 认证方案
```go
// 配置文件
auth:
  token: "your-secret-token-here"

// 中间件验证
Authorization: Bearer your-secret-token-here
```

### 6. 频率限制
- **策略**: Token Bucket算法
- **限制**: 每IP每分钟10次上传
- **响应**: 429 Too Many Requests

### 7. 内存缓存
```go
// 最近上传记录（内存中保存最近100条）
type RecentUploads struct {
    mu      sync.RWMutex
    uploads []UploadRecord
    maxSize int
}
```

## 前端功能

### 1. 上传方式
- **点击选择**: `<input type="file" accept="image/*" multiple>`
- **Ctrl+V粘贴**: 监听paste事件，读取clipboard
- **拖拽上传**: 监听drop事件
- **批量上传**: 支持多文件同时上传

### 2. 用户体验
- 上传前预览
- 实时进度条
- 上传成功后显示直链
- 一键复制链接
- 错误友好提示
- 响应式设计

### 3. 状态管理
```javascript
const [uploads, setUploads] = useState([])
const [uploading, setUploading] = useState(false)
const [progress, setProgress] = useState(0)
const [recentImages, setRecentImages] = useState([])
```

## Cloudflare CDN配置

### 1. Cache Rules
```
Cache Everything for:
- /images/*
- Cache TTL: 1 year
- Browser TTL: 1 year
```

### 2. 响应头
```go
w.Header().Set("Cache-Control", "public, max-age=31536000")
w.Header().Set("CDN-Cache-Control", "max-age=31536000")
```

### 3. 优化
- 启用Cloudflare Polish (图片优化)
- 启用Mirage (移动端优化)
- 启用WebP转换

## 构建流程

### Makefile
```makefile
.PHONY: build-frontend build-backend build clean

build: build-frontend build-backend

build-frontend:
	cd web && npm install && npm run build

build-backend:
	go build -o imgbed cmd/server/main.go

clean:
	rm -rf web/dist static imgbed
```

### 嵌入前端
```go
//go:embed static/*
var staticFS embed.FS

router.StaticFS("/", http.FS(staticFS))
```

## 配置文件

### config.yaml
```yaml
server:
  port: 8080
  host: 0.0.0.0

auth:
  token: "change-me-in-production"

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

## 安全考虑

1. **文件验证**: 魔数检查 + MIME类型 + 扩展名
2. **大小限制**: 5MB硬限制
3. **频率限制**: 防止滥用
4. **认证**: Bearer Token
5. **路径遍历**: 禁止../等危险字符
6. **CORS**: 配置允许的域名
7. **HTTPS**: 强制使用HTTPS

## 性能优化

1. **并发处理**: Goroutine处理图片
2. **内存优化**: 流式处理大文件
3. **CDN缓存**: Cloudflare缓存静态资源
4. **压缩**: Gzip响应压缩
5. **连接池**: 复用HTTP连接

## 监控和日志

```go
// 日志格式
{
  "timestamp": "2026-03-02T12:00:00Z",
  "level": "info",
  "action": "upload",
  "ip": "1.2.3.4",
  "filename": "abc123.jpg",
  "size": 1024000,
  "duration": "150ms"
}
```

## 部署

### 单二进制部署
```bash
# 构建
make build

# 运行
./imgbed

# 配置
export AUTH_TOKEN="your-secret-token"
./imgbed
```

### Docker部署
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /app/imgbed /imgbed
COPY config.yaml /config.yaml
EXPOSE 8080
CMD ["/imgbed"]
```

## 依赖包

### Go依赖
```
github.com/gin-gonic/gin
github.com/disintegration/imaging
github.com/google/uuid
github.com/spf13/viper
golang.org/x/time/rate
```

### 前端依赖
```
react
react-dom
axios
tailwindcss
vite
```
