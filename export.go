package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"textcleaner/internal/output"
)

// ExportFile 是桌面端原生导出的单个文件描述。
// Path 为相对目标目录的路径（可包含子目录），Data 为 base64 编码的文件内容。
type ExportFile struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

// ChooseDirectory 弹出系统“选择文件夹”对话框，返回所选目录（取消时返回空串）。
func (a *AppService) ChooseDirectory() (string, error) {
	if a.app == nil {
		return "", errors.New("应用未初始化")
	}
	result, err := a.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("选择导出文件夹").
		PromptForSingleSelection()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cancel") {
			return "", nil // 用户取消，视为无操作
		}
		return "", err
	}
	return result, nil
}

// ChooseSavePath 弹出系统“保存文件”对话框，返回所选路径（取消时返回空串）。
func (a *AppService) ChooseSavePath(defaultName string) (string, error) {
	if a.app == nil {
		return "", errors.New("应用未初始化")
	}
	result, err := a.app.Dialog.SaveFile().
		SetFilename(defaultName).
		PromptForSingleSelection()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cancel") {
			return "", nil
		}
		return "", err
	}
	return result, nil
}

// SaveFileBytes 将 base64 内容写入指定绝对路径（用于单文件 / ZIP 导出）。
// 目标路径由系统“保存文件”对话框返回，用户已确认可覆盖。
func (a *AppService) SaveFileBytes(path string, dataBase64 string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("保存路径为空")
	}
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return fmt.Errorf("解码内容失败: %w", err)
	}
	return output.SafeWrite(path, data, 0o644)
}

// ExportFiles 将多个处理结果写入用户选择的目标目录（多文件导出）。
// 每个文件的 Path 为相对目标目录的路径，Data 为 base64 内容。
// 返回成功写入的文件数量。
func (a *AppService) ExportFiles(dir string, files []ExportFile) (int, error) {
	root := output.Normalize(dir)
	if root == "" {
		return 0, errors.New("目标目录为空")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return 0, fmt.Errorf("创建目录失败: %w", err)
	}

	used := map[string]bool{}
	written := 0
	for _, f := range files {
		rel, err := sanitizeRelativePath(f.Path)
		if err != nil {
			return written, err
		}
		data, err := base64.StdEncoding.DecodeString(f.Data)
		if err != nil {
			return written, fmt.Errorf("解码 %q 失败: %w", f.Path, err)
		}
		dst := resolveUniquePath(root, rel, used)
		if err := output.SafeWrite(dst, data, 0o644); err != nil {
			return written, err
		}
		used[dst] = true
		written++
	}
	return written, nil
}

// sanitizeRelativePath 校验并规范化前端传入的相对路径，防止目录穿越和绝对路径。
func sanitizeRelativePath(p string) (string, error) {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if p == "" {
		return "", errors.New("文件路径为空")
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "/") {
		return "", fmt.Errorf("非法的绝对路径: %q", p)
	}
	// Windows 盘符（如 "C:/..." 或 "C:\\..."）
	if len(p) >= 2 && p[1] == ':' {
		return "", fmt.Errorf("非法的路径: %q", p)
	}
	clean := filepath.Clean(filepath.FromSlash(p))
	if clean == "." {
		clean = "unnamed.txt"
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("非法的路径（目录穿越）: %q", p)
	}
	return clean, nil
}

// resolveUniquePath 在 root 下为相对路径 rel 生成不与已用/已存在文件冲突的绝对路径。
func resolveUniquePath(root, rel string, used map[string]bool) string {
	dst := filepath.Join(root, rel)
	if !used[dst] && !pathExists(dst) {
		return dst
	}
	dir := filepath.Dir(dst)
	ext := filepath.Ext(dst)
	stem := strings.TrimSuffix(filepath.Base(dst), ext)
	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if !used[candidate] && !pathExists(candidate) {
			return candidate
		}
	}
	return dst
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
