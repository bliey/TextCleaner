package model

// FileEntry 表示一个已扫描到的文本文件。
type FileEntry struct {
	Path string `json:"path"` // 文件的绝对路径
	Name string `json:"name"` // 文件名（含扩展名）
	Size int64  `json:"size"` // 文件大小（字节）
	Ext  string `json:"ext"`  // 扩展名（含点，如 .txt）
}
