# 图床系统 - 前端

基于 React 18 + Vite + TailwindCSS 的现代化图床前端界面。

## 功能特性

- ✅ 用户认证（Bearer Token）
- ✅ 三种上传方式：
  - 点击选择文件
  - Ctrl+V 粘贴上传
  - 拖拽上传
- ✅ 批量上传支持
- ✅ 实时上传进度显示
- ✅ 图片预览（上传前后）
- ✅ 一键复制直链
- ✅ 最近上传历史
- ✅ 友好的错误提示
- ✅ 响应式设计（移动端适配）

## 技术栈

- React 18
- Vite 5
- TailwindCSS 3
- Axios

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问 http://localhost:5173
```

## 构建

```bash
# 构建生产版本（输出到 ../static 目录）
npm run build

# 预览构建结果
npm run preview
```

## 项目结构

```
web/
├── src/
│   ├── components/
│   │   ├── Auth.jsx           # 认证组件
│   │   ├── Upload.jsx         # 上传组件（支持三种上传方式）
│   │   ├── ImagePreview.jsx   # 图片预览组件
│   │   └── RecentUploads.jsx  # 最近上传列表
│   ├── App.jsx                # 主应用组件
│   ├── main.jsx               # 入口文件
│   └── index.css              # 全局样式
├── index.html
├── package.json
├── vite.config.js             # Vite 配置（输出到 static 目录）
└── tailwind.config.js         # TailwindCSS 配置
```

## API 集成

前端通过 axios 调用后端 API：

- `POST /api/upload` - 上传图片
- `GET /api/recent` - 获取最近上传

所有请求需要在 Header 中携带：
```
Authorization: Bearer <token>
```

## 构建配置

Vite 配置将构建产物输出到 `../static` 目录，供 Go 后端使用 `embed` 包嵌入到二进制文件中。
