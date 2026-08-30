package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"textcleaner/internal/model"
)

const fileName = "textcleaner-settings.json"

// Settings 需要持久化的用户偏好。
//
// 重要：只保存“勾选类”的通用选项（如基础清理开关、默认包含子文件夹、
// 并发数、主题、语言）。用户每次输入的具体删除 / 替换文本规则不在此保存。
type Settings struct {
	BasicClean             model.BasicCleanOptions `json:"basicClean"`
	DefaultIncludeSubfolders bool                  `json:"defaultIncludeSubfolders"`
	MaxConcurrency         int                     `json:"maxConcurrency"`
	Theme                  string                  `json:"theme"`  // light | dark | system
	Language               string                  `json:"language"` // auto | zh | en
}

// DefaultSettings 返回出厂默认设置（符合产品建议的勾选项）。
func DefaultSettings() Settings {
	return Settings{
		BasicClean: model.BasicCleanOptions{
			TrimLeadingWhitespace:  true,
			TrimTrailingWhitespace: true,
			RemoveZeroWidthChars:   true,
			CollapseBlankLines:     true,
			MaxBlankLines:          1,
		},
		DefaultIncludeSubfolders: true,
		MaxConcurrency:           4,
		Theme:                    "system",
		Language:                 "auto",
	}
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(dir, "TextCleaner")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(base, fileName), nil
}

// Load 读取持久化偏好；文件不存在时返回默认设置。
func Load() (Settings, error) {
	def := DefaultSettings()
	path, err := settingsPath()
	if err != nil {
		return def, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return def, nil
		}
		return def, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return def, err
	}
	return s, nil
}

// Save 保存偏好到用户配置目录。
func Save(s Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
