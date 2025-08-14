# 文件上传系统前端

这是一个基于Vue 3和Element Plus的文件上传系统前端，支持大文件上传和断点续传功能。

## 功能特点

- 大文件上传：支持上传任意大小的文件
- 断点续传：支持暂停和恢复上传
- 文件分片：将大文件分割成小块进行上传
- 文件哈希：使用MD5计算文件哈希，支持秒传
- 上传进度：实时显示上传进度和分片状态
- 文件管理：支持文件列表查看、下载和删除

## 技术栈

- Vue 3：前端框架
- Element Plus：UI组件库
- Axios：HTTP请求库
- SparkMD5：文件哈希计算

## 项目结构

```
front/
├── public/              # 静态资源
├── src/
│   ├── assets/          # 资源文件
│   ├── components/      # 公共组件
│   ├── utils/           # 工具函数
│   │   ├── api.js       # API请求
│   │   ├── file.js      # 文件处理工具
│   │   └── request.js   # Axios配置
│   ├── views/           # 页面组件
│   │   ├── FileList.vue # 文件列表页面
│   │   └── FileUpload.vue # 文件上传页面
│   ├── App.vue          # 根组件
│   ├── main.js          # 入口文件
│   └── router/          # 路由配置
├── package.json         # 项目依赖
└── vue.config.js        # Vue配置
```

## 开发指南

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run serve
```

### 构建生产版本

```bash
npm run build
```

## 使用说明

1. 在上传页面选择或拖拽文件
2. 系统会自动计算文件哈希
3. 点击"开始上传"按钮开始上传
4. 上传过程中可以暂停和恢复
5. 上传完成后可以在文件列表页面查看和管理文件