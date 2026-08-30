package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails 使用 Go 的 embed 包把前端构建产物（frontend/dist）打包进二进制。
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc := &AppService{}

	app := application.New(application.Options{
		Name:        "text-cleaner",
		Description: "批量文本清理与替换工具",
		// Services 中列出的结构体，其导出方法会被暴露给前端调用。
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 让服务能主动向前端推送事件（如批量处理进度）。
	svc.app = app

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Text Cleaner",
		Width:            1280,
		Height:           860,
		MinWidth:         920,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(246, 247, 249),
		URL:              "/",
		// 允许把文件 / 文件夹拖入窗口（前端元素需带 data-file-drop-target）
		EnableFileDrop: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	// 文件拖放：把窗口级事件转发为前端可监听的自定义事件。
	// 路径解析在 Go 端完成，前端只负责接收路径并展示。
	win.OnWindowEvent(events.Common.WindowFilesDropped, func(e *application.WindowEvent) {
		files := e.Context().DroppedFiles()
		if len(files) > 0 {
			app.Event.Emit("files-dropped", files)
		}
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
