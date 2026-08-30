package processor

import (
	"regexp"

	"textcleaner/internal/model"
)

// applyRegex 应用正则替换规则（按顺序逐条执行）。
// 普通文本规则已由 applyReplace 处理，这里跳过。
//
// 当某条规则的 Replace 为空字符串时，视为“删除”操作，
// 其匹配数计入 deleted；否则计入 replaced。
func applyRegex(text string, rules []model.ReplaceRule) (string, int, int) {
	deleted := 0
	replaced := 0
	for _, r := range rules {
		if !r.Enabled || !r.Regex {
			continue
		}
		if r.Find == "" {
			continue
		}
		re, err := regexp.Compile(r.Find)
		if err != nil {
			// 非法正则跳过（前端应在提交前校验，这里仅做防御）
			continue
		}
		matches := re.FindAllString(text, -1)
		if len(matches) > 0 {
			text = re.ReplaceAllString(text, r.Replace)
			if r.Replace == "" {
				deleted += len(matches)
			} else {
				replaced += len(matches)
			}
		}
	}
	return text, deleted, replaced
}
