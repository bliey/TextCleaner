package model

// LineEnding 指定统一换行符的目标。
type LineEnding string

const (
	LineEndingKeep LineEnding = "keep" // 保持原文件换行符
	LineEndingLF   LineEnding = "lf"   // 统一为 LF (\n)
	LineEndingCRLF LineEnding = "crlf" // 统一为 CRLF (\r\n)
)
