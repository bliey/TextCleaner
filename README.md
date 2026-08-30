# TextCleaner

在线批量文本清理工具，所有处理均在浏览器本地完成。

## 在线使用

https://bliey.github.io/TextCleaner/

## 功能

* 批量导入 `.txt`、`.md`、`.log`、`.csv`
* 支持文件和文件夹导入
* 支持拖拽
* 保留文件夹结构
* 文本清理与固定文本删除
* 批量替换与正则替换
* 简体 / 繁体转换
* 支持 UTF-8、UTF-16、GBK / GB18030、Big5
* 实时预览与差异对比
* 处理进度与取消
* ZIP 导出

## 隐私

所有文件均在浏览器本地处理，不会上传到服务器。

## 本地开发

```bash
cd frontend
npm install
npm run dev
```

构建：

```bash
npm run build
```

推送到 `main` 后，GitHub Actions 会自动部署到 GitHub Pages。
