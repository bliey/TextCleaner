package processor

import (
	"strings"

	"textcleaner/internal/model"
)

// applyReplace 应用普通文本替换规则（按顺序逐条执行）。
// 正则规则交由 applyRegex 处理，这里跳过。
//
// 当某条规则的 Replace 为空字符串时，视为“删除”操作，
// 其匹配数计入 deleted；否则计入 replaced。
func applyReplace(text string, rules []model.ReplaceRule) (string, int, int) {
	deleted := 0
	replaced := 0
	for _, r := range rules {
		if !r.Enabled || r.Regex {
			continue
		}
		if r.Find == "" {
			continue
		}
		n := strings.Count(text, r.Find)
		if n > 0 {
			text = strings.ReplaceAll(text, r.Find, r.Replace)
			if r.Replace == "" {
				deleted += n
			} else {
				replaced += n
			}
		}
	}
	return text, deleted, replaced
}
