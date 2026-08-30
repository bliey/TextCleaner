package model

// ReplaceRule 一条替换规则（普通文本或正则）。
//
// “删除”只是“替换为空白内容”的特例：当 Replace 为空字符串时，
// 该规则的效果即删除所有匹配到的 Find。因此删除与替换共用同一套规则，
// 不再单独存在 Delete 字段。
type ReplaceRule struct {
	Enabled bool   `json:"enabled"`
	Find    string `json:"find"`
	Replace string `json:"replace"`
	Regex   bool   `json:"regex"`
}

// BasicCleanOptions “基础功能”区的全部开关选项（合并后）。
//
// 历史上 BasicCleaner 9 条 + ExtraCleaner 17 条共 26 条规则分散在两个
// struct 里、UI 上也是两个 SectionCard。经用户要求统一为单一「基础功能」
// 分区后，本 struct 收纳所有字段；json tag 保持 "basicClean" 不变，
// 以保证老版本持久化的 settings.json 仍可读取（缺的字段为零值）。
//
// 设计要点：
//   - 全部为 bool 开关，单条规则独立可勾选；
//   - 字段名即 Go 端约定的开关常量，前端 binding 自动生成同名 TypeScript 字段；
//   - 命名风格为 UpperCamelCase，对应前端 lowerCamelCase（binding 自动转换）；
//   - 各规则的执行顺序由 cleanrules 包内部编排，此处不关心。
type BasicCleanOptions struct {
	// ---- 字符（Character-level normalization） ----
	TrimLeadingWhitespace  bool `json:"trimLeadingWhitespace"`  // 删除行首空白
	TrimTrailingWhitespace bool `json:"trimTrailingWhitespace"` // 删除行尾空白
	RemoveUTF8BOM          bool `json:"removeUTF8BOM"`          // 删除 UTF-8 BOM
	RemoveZeroWidthChars   bool `json:"removeZeroWidthChars"`   // 删除零宽字符
	FullWidthToHalfWidth   bool `json:"fullWidthToHalfWidth"`   // 全角字符转半角字符
	ReplaceKangxiRadicals  bool `json:"replaceKangxiRadicals"`  // 康熙部首替换为正常字符

	// ---- 标点（Punctuation conversion） ----
	PunctEnglishToChinese bool `json:"punctEnglishToChinese"` // 英文标点 → 中文标点
	PunctChineseToEnglish bool `json:"punctChineseToEnglish"` // 中文标点 → 英文标点

	// ---- 引用角标（Citation marks） ----
	RemoveCiteParen   bool `json:"removeCiteParen"`   // 删除 (1)、(2,3)、(4-7)
	RemoveCiteBracket bool `json:"removeCiteBracket"` // 删除 [1]、【2,3】、[4-7]

	// ---- 空白与换行（Whitespace / newlines） ----
	CollapseNewlines bool `json:"collapseNewlines"` // 合并连续换行为单个（严格）
	NewlineToSpace   bool `json:"newlineToSpace"`   // 所有换行 → 空格
	CollapseSpaces   bool `json:"collapseSpaces"`   // 合并连续空格为单个
	RemoveEmptyLines bool `json:"removeEmptyLines"` // 删除完全空白的行
	CollapseBlankLines bool `json:"collapseBlankLines"` // 合并连续空行至最多 MaxBlankLines
	MaxBlankLines      int  `json:"maxBlankLines"`       // 上述合并后保留的最大连续空行数

	// ---- CJK / 数字 间距 ----
	RemoveSpaceBetweenCJK      bool `json:"removeSpaceBetweenCJK"`      // 删除 CJK 字符之间的空格
	SpaceAfterPunctuation      bool `json:"spaceAfterPunctuation"`      // 在标点符号后添加空格
	RemoveSpaceAtDecimal       bool `json:"removeSpaceAtDecimal"`       // 删除小数点与数字之间的空格
	SpaceBetweenLetterAndDigit bool `json:"spaceBetweenLetterAndDigit"` // 在字母与数字之间添加空格
	RemoveSpaceAtColon         bool `json:"removeSpaceAtColon"`         // 删除冒号与数字之间的空格

	// ---- 中文排版（参考 sspai.com/post/37815） ----
	NormalizeChineseTypography bool `json:"normalizeChineseTypography"` // 规范中文排版

	// ---- 简繁转换 ----
	SimplifiedToTraditional bool `json:"simplifiedToTraditional"` // 简 → 繁
	TraditionalToSimplified bool `json:"traditionalToSimplified"` // 繁 → 简

	// ---- 输出换行符（仅在 Clean 时影响是否统一；最终 LF/CRLF/keep 由 processor.toEnding 决定） ----
	NormalizeLineEndings bool       `json:"normalizeLineEndings"` // 是否在 Clean 阶段先归一为 LF
	LineEnding           LineEnding `json:"lineEnding"`           // 最终输出换行符（NormalizeLineEndings 为真时生效）
}

// ProcessOptions 一次文本处理的完整配置。
//
// 注意：删除内容已合并进 Replace（Replace 为空即删除），不再单独成字段。
// BasicClean 收纳所有“基础功能”开关（共 26 项），由 cleanrules.ApplyExtra 执行。
type ProcessOptions struct {
	BasicClean BasicCleanOptions `json:"basicClean"`
	Replace    []ReplaceRule     `json:"replace"`
}
