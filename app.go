package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"textcleaner/internal/batch"
	"textcleaner/internal/config"
	"textcleaner/internal/encoding"
	"textcleaner/internal/model"
	"textcleaner/internal/processor"
	"textcleaner/internal/scanner"
)

// AppService 暴露给前端（Vue）的后端服务。
//
// 设计原则：所有核心文本处理都由 Go 完成，前端只负责展示与交互，
// 不重复实现任何文本处理逻辑。
type AppService struct {
	app *application.App

	mu     sync.Mutex
	cancel context.CancelFunc
}

// Version 返回应用版本，用于前端展示（同时验证 Go ↔ Vue 调用链路）。
func (a *AppService) Version() string {
	return "0.1.0"
}

// ProcessText 文本处理核心入口。预览与批量处理共用此函数，确保逻辑一致。
func (a *AppService) ProcessText(input string, options model.ProcessOptions) (model.ProcessOutput, error) {
	text, res, err := processor.ProcessText(input, options)
	if err != nil {
		return model.ProcessOutput{}, err
	}
	return model.ProcessOutput{Text: text, Result: res}, nil
}

// GetSettings 读取持久化的偏好设置。
func (a *AppService) GetSettings() (config.Settings, error) {
	return config.Load()
}

// SaveSettings 保存偏好设置。
func (a *AppService) SaveSettings(s config.Settings) error {
	return config.Save(s)
}

// ScanPaths 扫描给定的路径（可混合文件与文件夹），返回去重、排序后的文本文件列表。
// 文件系统遍历统一在 Go 端完成，前端不重复实现。
func (a *AppService) ScanPaths(paths []string, options scanner.Options) ([]model.FileEntry, error) {
	return scanner.ScanPaths(paths, options)
}

// ReadTextFile 读取文本文件的原始字节并以字符串返回（按编码自动识别解码）。
func (a *AppService) ReadTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text, _, _ := encoding.Decode(data)
	return text, nil
}

// ProcessBatch 批量处理文件。该方法立即返回，进度通过 "batch-progress" 事件推送，
// 完成时通过 "batch-done" 事件推送汇总。可通过 CancelBatch 取消。
func (a *AppService) ProcessBatch(opts batch.Options) error {
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return errors.New("已有批量任务正在进行中，请先取消")
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.mu.Unlock()

	go func() {
		summary, err := batch.Run(opts, ctx, func(p batch.Progress) {
			if a.app != nil {
				a.app.Event.Emit("batch-progress", p)
			}
		})

		a.mu.Lock()
		a.cancel = nil
		a.mu.Unlock()

		if a.app != nil {
			if err != nil {
				a.app.Event.Emit("batch-done", map[string]any{"error": err.Error(), "summary": summary})
			} else {
				a.app.Event.Emit("batch-done", map[string]any{"summary": summary})
			}
		}
	}()
	return nil
}

// CancelBatch 取消正在进行的批量任务。
func (a *AppService) CancelBatch() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

// OpenFolder 在系统文件管理器中打开指定目录（跨平台）。
// 批量处理完成后，前端用它让用户一键跳转到输出目录。
func (a *AppService) OpenFolder(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
