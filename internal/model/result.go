package model

// ProcessResult 单次 ProcessText 的统计信息。
type ProcessResult struct {
	DeletedMatches  int  `json:"deletedMatches"`  // 被删除的匹配数量
	ReplacedMatches int  `json:"replacedMatches"` // 被替换的匹配数量
	Changed         bool `json:"changed"`         // 文本是否发生了变化
}

// ProcessOutput 文本处理输出，供前端预览 / 批量处理使用。
type ProcessOutput struct {
	Text   string       `json:"text"`   // 处理后的文本
	Result ProcessResult `json:"result"` // 处理统计
}

// FileProcessResult 单个文件批量处理的结果。
type FileProcessResult struct {
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
