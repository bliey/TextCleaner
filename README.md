# TextCleaner

批量文本清理与替换工具，支持 Windows / macOS 桌面端与 Web 网页版。所有文本处理均在本地完成，不会上传任何数据。

## 平台支持

| 平台 | 交付物 | 获取方式 |
| --- | --- | --- |
| Windows (x64) | 安装程序 `.exe` | [Releases](https://github.com/bliey/TextCleaner/releases) |
| macOS (Apple Silicon arm64) | `.app`（zip 压缩包） | [Releases](https://github.com/bliey/TextCleaner/releases) |
| macOS (Intel x64) | `.app`（zip 压缩包） | [Releases](https://github.com/bliey/TextCleaner/releases) |
| Web | 在线使用 | https://bliey.github.io/TextCleaner/ |

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
* 桌面版（Windows/macOS）原生导出：系统保存对话框 / 选择文件夹直接写入

## 构建

项目基于 [Wails v3](https://wails.io)（Go 后端壳）+ Vue 3 / Vite / TypeScript（前端），使用 [go-task](https://taskfile.dev) 作为构建编排器。

前置要求：

* Go 1.25+
* Node.js 20+
* `wails3` CLI（版本与 `go.mod` 一致）：`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.15`
* `task` CLI：`go install github.com/go-task/task/v3/cmd/task@latest`

### Web（构建静态站点）

```sh
cd frontend
npm ci
npm run build
# 产物输出到 frontend/dist/
```

GitHub Pages 部署时由 CI 设置 `VITE_BASE=/TextCleaner/`。

### Windows（构建主程序 + NSIS 安装包）

```sh
task build                 # 生成 bin/text-cleaner.exe（需先构建前端，任务会自动处理）
task package               # 生成 NSIS 安装包 bin/text-cleaner-amd64-installer.exe（需安装 NSIS）
```

### macOS（构建 .app bundle）

需在 macOS 上执行：

```sh
task build ARCH=arm64      # 或 ARCH=amd64，生成 bin/text-cleaner
task package ARCH=arm64    # 生成 bin/text-cleaner.app
```

发布时由 GitHub Actions 在 `macos-14`（arm64）与 `macos-13`（x64）runner 上分别构建并打包。

### 开发

```sh
task dev                   # Wails 开发模式（热重载）
# 或只跑前端
cd frontend && npm run dev
```

## 隐私

所有文件均在本地处理，不会上传到服务器。
