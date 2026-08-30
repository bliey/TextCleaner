# TextCleaner

在线文本批量清理工具，所有文本处理均在浏览器本地完成。

## 在线使用

<https://bliey.github.io/TextCleaner/>

## 功能

- 批量导入 `.txt`、`.md`、`.log`、`.csv`
- 选择多个文件或文件夹
- 文件夹与子目录结构保留到 ZIP
- 文件拖拽导入
- 基础文本清理
- 固定文本删除与批量替换
- 正则表达式替换
- 简体 / 繁体转换
- UTF-8、UTF-16、GBK / GB18030、Big5 文本读取
- 实时预览与差异对比
- 真实处理进度与取消
- ZIP 导出与文件名冲突自动编号

## 隐私

所有文件均在您的浏览器本地处理，不会上传到服务器。GitHub Pages 只负责托管静态 HTML、CSS 和 JavaScript。

## 开发

Web 前端位于 `frontend/`：

```bash
cd frontend
npm install
npm run dev
```

生产构建：

```bash
cd frontend
npm run build
```

构建产物位于 `frontend/dist/`。推送到 `main` 后，GitHub Actions 会自动构建并部署到 GitHub Pages。

## 迁移说明

仓库保留了原 Wails / Go 桌面版源码，便于对照验证；GitHub Pages 只构建 `frontend/`，运行时不依赖 Go、Wails、Node.js 服务、数据库或任何后端 API。Web 版本使用 Browser File API、TypeScript 文本处理核心和 JSZip 完成文件处理与导出。
