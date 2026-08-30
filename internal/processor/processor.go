package processor

import (
	"strings"

	"textcleaner/internal/cleanrules"
	"textcleaner/internal/model"
)

// ProcessText 文本处理核心入口。
//
// 批量文件处理与预览都调用此函数，确保二者逻辑完全一致。
//
// 处理顺序（合并“基础清理”与“基础功能”为单一 BasicCleanOptions 后）：
//
//	① 内部统一为 \n
//	② 全部基础功能开关（cleanrules.ApplyExtra，含字符规整、引用角标删除、
//	   空白与换行整理、CJK 间距、标点/排版、简繁转换等 26 项）
//	③ 用户自定义的删除/替换/正则规则（按顺序应用）
//	④ 决定最终换行符（LF / CRLF / keep），保持末尾换行（若原文件以 \n 结尾）
//
// 26 项开关的具体编排见 internal/cleanrules/cleanrules.go 顶部注释。
//
// 编码 / 解码由调用方（文件读写层）负责，本函数只处理字符串。
func ProcessText(input string, opt model.ProcessOptions) (string, model.ProcessResult, error) {
	// ① 内部统一以 \n 处理
	text := toLF(input)

	// ② pre-clean：BOM / 行首尾空白 / 零宽 / 连续空格（这些在用户规则之前跑，
	//    让用户的 find 字符串能稳定命中，与源文件的格式怪癖解耦）
	text = cleanrules.PreClean(text, opt.BasicClean)

	// ③ 用户自定义的删除 / 替换 / 正则（按顺序逐条执行；空 Replace 视为删除）
	var delNormal, delRegex, repNormal, repRegex int
	text, delNormal, repNormal = applyReplace(text, opt.Replace)
	text, delRegex, repRegex = applyRegex(text, opt.Replace)

	// ④ post-clean：其余 22 项基础功能开关（字符规整 → 引用角标 → 空白换行
	//    → CJK 间距 → 标点/排版 → 简繁 → 空行整理）；放在用户规则之后，使得
	//    用户删掉内容后产生的空行能被合并，用户新增的标点能被规整
	text = cleanrules.ApplyExtra(text, opt.BasicClean)

	// ⑤ 决定最终换行符
	ending := detectLineEnding(input)
	if opt.BasicClean.NormalizeLineEndings {
		switch opt.BasicClean.LineEnding {
		case model.LineEndingCRLF:
			ending = model.LineEndingCRLF
		case model.LineEndingLF:
			ending = model.LineEndingLF
		default:
			// keep：沿用原文件换行符
			ending = detectLineEnding(input)
		}
	}
	final := toEnding(text, ending)

	// 保留原文件的结尾换行符（若原本以换行结尾，且结果非空且不以其结尾，则补回）。
	if strings.HasSuffix(toLF(input), "\n") && final != "" {
		sep := "\n"
		if ending == model.LineEndingCRLF {
			sep = "\r\n"
		}
		if !strings.HasSuffix(final, sep) {
			final += sep
		}
	}

	res := model.ProcessResult{
		DeletedMatches:  delNormal + delRegex,
		ReplacedMatches: repNormal + repRegex,
		Changed:         final != input,
	}
	return final, res, nil
}

// ============================================================
// Line-ending helpers (moved here from the now-deleted cleaner.go,
// which used to host preClean / postClean that have been merged
// into cleanrules.ApplyExtra).
// ============================================================

// toLF 将文本统一转换为以 \n 分隔（便于内部按行处理）。
func toLF(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// toEnding 按目标换行符重新拼接文本。
func toEnding(s string, ending model.LineEnding) string {
	if ending == model.LineEndingCRLF {
		return strings.ReplaceAll(s, "\n", "\r\n")
	}
	return s
}

// detectLineEnding 推断文本主要使用的换行符。
func detectLineEnding(s string) model.LineEnding {
	if strings.Contains(s, "\r\n") {
		return model.LineEndingCRLF
	}
	if strings.Contains(s, "\r") {
		return model.LineEndingCRLF // 旧版 Mac (\r) 统一按 CRLF 处理
	}
	return model.LineEndingLF
}
