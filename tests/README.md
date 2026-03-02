# 图床系统测试文档

## 测试概述

本测试套件为图床系统提供全面的测试覆盖，包括单元测试、集成测试、E2E测试、安全测试和性能测试。

## 测试文件结构

```
tests/
├── image_test.go      # 图片处理单元测试
├── api_test.go        # API集成测试
├── storage_test.go    # 存储功能测试
└── e2e_test.go        # 端到端测试
```

## 运行测试

### 运行所有测试
```bash
go test ./tests/... -v
```

### 运行特定测试文件
```bash
go test ./tests/image_test.go -v
go test ./tests/api_test.go -v
go test ./tests/storage_test.go -v
go test ./tests/e2e_test.go -v
```

### 运行特定测试用例
```bash
go test ./tests/... -v -run TestValidator_FileSizeLimit
go test ./tests/... -v -run TestAPI_Upload_Success
```

### 运行性能测试
```bash
go test ./tests/... -bench=. -benchmem
```

### 生成测试覆盖率报告
```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 测试覆盖范围

### 1. 图片处理测试 (image_test.go)

#### 文件验证测试
- **TestValidator_FileSizeLimit**: 测试文件大小限制（5MB）
- **TestValidator_FormatValidation**: 测试文件格式验证（JPEG/PNG/GIF/WebP）
- **TestValidator_MagicNumberCheck**: 测试魔数检查，防止文件伪造
- **TestValidator_GetExtension**: 测试扩展名获取

#### 图片处理测试
- **TestProcessor_ImageResizing**: 测试图片缩放（>2000px自动缩放）
- **TestProcessor_ThumbnailGeneration**: 测试缩略图生成（300x300）
- **BenchmarkProcessor_Process**: 图片处理性能基准测试

#### 存储测试
- **TestStorage_DirectoryCreation**: 测试目录自动创建（YYYY/MM结构）
- **TestStorage_FilenameFormat**: 测试文件命名格式（YYYYMMDDhhmmss_uuid.ext）
- **TestStorage_CleanupOnFailure**: 测试失败时的清理机制

### 2. API集成测试 (api_test.go)

#### 认证测试
- **TestAPI_Upload_MissingAuth**: 测试缺少认证令牌
- **TestAPI_Upload_InvalidToken**: 测试无效认证令牌
- **TestAuth_ValidateToken**: 测试Bearer Token验证

#### 上传功能测试
- **TestAPI_Upload_NoFile**: 测试未上传文件的错误处理
- **TestAPI_Upload_FileTooLarge**: 测试超过5MB限制的文件
- **TestAPI_Upload_InvalidFormat**: 测试无效文件格式
- **TestAPI_Upload_Success**: 测试成功上传流程
- **TestAPI_Upload_MultipleFormats**: 测试多种图片格式支持

#### API端点测试
- **TestAPI_HealthCheck**: 测试健康检查端点
- **TestAPI_RecentUploads_Success**: 测试获取最近上传记录
- **TestAPI_ServeImage_Success**: 测试图片访问和CDN缓存头
- **TestAPI_ServeImage_NotFound**: 测试不存在的图片

#### 频率限制测试
- **TestAPI_RateLimit**: 测试IP频率限制（10次/分钟）
- **TestRateLimit_IPLimiter**: 测试Token Bucket算法实现

#### 并发测试
- **TestAPI_ConcurrentUploads**: 测试并发上传处理
- **BenchmarkAPI_Upload**: API上传性能基准测试

#### 完整流程测试
- **TestAPI_UploadAndRetrieveFlow**: 测试上传→获取→最近记录完整流程

### 3. 存储功能测试 (storage_test.go)

#### 基础存储测试
- **TestStorage_BasicSave**: 测试基本保存操作
- **TestStorage_DirectoryStructure**: 测试目录结构（uploads/YYYY/MM/）
- **TestStorage_FilenameFormat**: 测试文件名格式验证
- **TestStorage_MultipleExtensions**: 测试多种扩展名支持

#### 高级功能测试
- **TestStorage_ThumbnailNaming**: 测试缩略图命名（_thumb后缀）
- **TestStorage_FilePermissions**: 测试文件权限（0755）
- **TestStorage_UniqueFilenames**: 测试文件名唯一性
- **TestStorage_QualitySettings**: 测试不同质量设置（50/75/90/95）

#### 并发和性能测试
- **TestStorage_ConcurrentSaves**: 测试并发保存操作
- **TestStorage_LargeImage**: 测试大图片处理（4000x3000）
- **BenchmarkStorage_Save**: 存储性能基准测试
- **BenchmarkStorage_ConcurrentSave**: 并发存储性能测试

#### 安全测试
- **TestStorage_PathTraversalPrevention**: 测试路径遍历防护
- **TestStorage_GetImagePath**: 测试路径构建安全性

### 4. 端到端测试 (e2e_test.go)

#### 完整流程测试
- **TestE2E_CompleteUploadFlow**: 测试完整上传流程
  1. 健康检查
  2. 上传图片
  3. 验证磁盘文件
  4. 通过API获取图片
  5. 检查最近上传记录

#### 多场景测试
- **TestE2E_MultipleImageUpload**: 测试批量上传（5张图片）
- **TestE2E_ErrorHandlingFlow**: 测试各种错误场景
- **TestE2E_ImageSizeHandling**: 测试不同尺寸图片处理

#### 压力测试
- **TestE2E_ConcurrentUploadStress**: 并发上传压力测试（10个并发）
- **TestE2E_RateLimitingBehavior**: 测试频率限制行为（20次请求）

#### 功能验证测试
- **TestE2E_ImageFormatSupport**: 测试多种图片格式支持
- **TestE2E_CacheHeaders**: 测试CDN缓存头设置
- **TestE2E_FullWorkflowWithVerification**: 完整工作流验证

## 测试场景详解

### 安全测试场景

1. **文件验证**
   - 魔数检查：防止通过修改扩展名伪造文件类型
   - MIME类型验证：检查Content-Type头
   - 文件大小限制：硬限制5MB

2. **认证测试**
   - Bearer Token验证
   - 无效令牌拒绝
   - 缺少认证头拒绝

3. **频率限制**
   - IP级别限流（10次/分钟）
   - Token Bucket算法
   - 不同IP独立限制

4. **路径安全**
   - 防止路径遍历攻击
   - 文件名安全验证

### 性能测试场景

1. **图片处理性能**
   - 800x600图片处理基准
   - 大图片（4000x3000）处理
   - 缩略图生成性能

2. **并发处理**
   - 10个并发上传
   - 并发存储操作
   - 并发API请求

3. **存储性能**
   - 不同质量设置的文件大小
   - 磁盘I/O性能
   - 目录创建性能

### 边界测试场景

1. **文件大小边界**
   - 0字节文件
   - 5MB限制边界
   - 超大文件（6MB）

2. **图片尺寸边界**
   - 小图片（100x100）
   - 临界尺寸（2000x1500）
   - 超大尺寸（3000x2000）

3. **并发边界**
   - 单个请求
   - 5个并发（burst限制）
   - 10个并发
   - 20个请求（触发限流）

## 测试数据

### 测试图片生成
测试使用程序生成的图片，包含渐变色彩，确保：
- 可压缩性测试
- 不同尺寸测试
- 不同格式测试

### 测试配置
```go
MaxSize:      5 * 1024 * 1024  // 5MB
AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
MaxDimension: 2000
Quality:      90
ThumbnailSize: 300
RequestsPerMinute: 10
Burst: 5
```

## 预期测试结果

### 成功标准
- 所有单元测试通过
- API集成测试通过
- E2E测试完整流程通过
- 无内存泄漏
- 并发测试无竞态条件

### 性能指标
- 单张图片上传处理 < 500ms
- 并发10个请求全部成功
- 内存使用稳定
- 无goroutine泄漏

## 故障排查

### 常见测试失败原因

1. **依赖未安装**
   ```bash
   go mod tidy
   go mod download
   ```

2. **临时目录权限**
   - 确保测试有权限创建临时目录
   - 检查磁盘空间

3. **端口占用**
   - 测试使用httptest，不需要真实端口

4. **网络超时**
   - 测试不依赖外部网络
   - 使用本地mock数据

## 持续集成

### GitHub Actions示例
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go mod download
      - run: go test ./tests/... -v -race -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out
```

## 测试维护

### 添加新测试
1. 在相应的测试文件中添加测试函数
2. 使用`Test`前缀命名
3. 使用`t.Run()`创建子测试
4. 添加清晰的测试描述

### 测试最佳实践
- 使用表驱动测试
- 每个测试独立运行
- 清理测试数据（使用defer）
- 使用有意义的测试名称
- 添加测试日志（t.Logf）

## 总结

本测试套件提供了全面的测试覆盖：
- ✅ 单元测试：图片处理、验证、存储
- ✅ 集成测试：API端点、认证、限流
- ✅ E2E测试：完整上传流程
- ✅ 安全测试：文件验证、认证、限流
- ✅ 性能测试：并发处理、大文件
- ✅ 边界测试：文件大小、格式、尺寸

测试覆盖了所有关键功能和边界情况，确保系统的稳定性和安全性。
