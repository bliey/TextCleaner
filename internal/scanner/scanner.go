package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"textcleaner/internal/model"
)

// maxFilesHardLimit 是防误操作的保护上限：避免误选超大磁盘目录导致
// 内存暴涨或 UI 失去响应。用户可在 Options.MaxFiles 中设置更小的上限。
const maxFilesHardLimit = 500000

// DefaultExtensions 是第一阶段重点支持的纯文本扩展名。
// 仅作为界面默认值，不会硬编码到处理逻辑中。
var DefaultExtensions = []string{".txt", ".md", ".log", ".csv"}

// Options 扫描选项。
type Options struct {
	IncludeSubfolders bool     `json:"includeSubfolders"` // 是否递归扫描子文件夹
	Extensions        []string `json:"extensions"`        // 允许的扩展名（.txt / txt / *.txt 均可）；为空表示不限类型
	MaxFiles          int      `json:"maxFiles"`          // 最大文件数；<=0 表示使用内置硬上限
}

// ScanPaths 扫描一组路径（可混合文件与文件夹），返回去重并排序后的文件条目。
//
// 使用 filepath.WalkDir 流式遍历，不会一次性把目录结构全部载入内存；
// 单个路径不可访问时跳过而不中断整体扫描。
func ScanPaths(paths []string, opt Options) ([]model.FileEntry, error) {
	exts := normalizeExts(opt.Extensions)
	limit := opt.MaxFiles
	if limit <= 0 || limit > maxFilesHardLimit {
		limit = maxFilesHardLimit
	}

	seen := make(map[string]struct{})
	entries := make([]model.FileEntry, 0, 256)

	for _, raw := range paths {
		p := filepath.Clean(strings.TrimSpace(raw))
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			// 路径不存在或无权访问：跳过，不中断其余路径
			continue
		}
		if info.IsDir() {
			if err := scanDir(p, opt.IncludeSubfolders, exts, limit, seen, &entries); err != nil {
				return nil, err
			}
			continue
		}
		if matchesExt(p, exts) {
			addEntry(p, info, seen, &entries)
		}
		if len(entries) >= limit {
			break
		}
	}

	// 按路径排序，保证结果稳定可预期
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

// scanDir 遍历单个目录；includeSub 为 false 时不进入子目录。
func scanDir(root string, includeSub bool, exts []string, limit int, seen map[string]struct{}, out *[]model.FileEntry) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 权限不足等：跳过该条目，继续扫描
			return nil
		}
		if d.IsDir() {
			// 不递归子文件夹：除根目录本身外一律跳过
			if !includeSub && path != root {
				return fs.SkipDir
			}
			return nil
		}
		if !matchesExt(path, exts) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		addEntry(path, info, seen, out)
		if len(*out) >= limit {
			return filepath.SkipAll
		}
		return nil
	})
}

// addEntry 追加一个文件条目并去重。
func addEntry(path string, info fs.FileInfo, seen map[string]struct{}, out *[]model.FileEntry) {
	if _, ok := seen[path]; ok {
		return
	}
	seen[path] = struct{}{}
	*out = append(*out, model.FileEntry{
		Path: path,
		Name: info.Name(),
		Size: info.Size(),
		Ext:  strings.ToLower(filepath.Ext(path)),
	})
}

// normalizeExts 把 "txt" / ".txt" / "*.txt" 统一规范为小写的 ".txt"。
func normalizeExts(in []string) []string {
	out := make([]string, 0, len(in))
	for _, raw := range in {
		e := strings.TrimSpace(raw)
		e = strings.TrimPrefix(e, "*")
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out = append(out, strings.ToLower(e))
	}
	return out
}

// matchesExt 判断文件扩展名是否命中；exts 为空时表示不做限制。
func matchesExt(path string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}
