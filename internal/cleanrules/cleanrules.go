// Package cleanrules implements the 26 "基础功能" toggleable text-cleaning
// rules exposed in the UI's Basic Features section. Each rule is a small
// pure function. The package exposes two entry points to keep ordering
// with the user's find/replace rules correct:
//
//   - PreClean(text, opt): line-level normalization that must run BEFORE
//     user rules (BOM, per-line trim, zero-width strip, collapse-spaces).
//     Running these first lets user-supplied find strings match
//     reliably regardless of source formatting quirks.
//   - ApplyExtra(text, opt): the remaining 22 rules — character-level
//     normalization (full-width→half-width, Kangxi), citation-mark
//     deletion, whitespace/newline management, CJK spacing,
//     punctuation conversion, Chinese typography, S↔T, and
//     blank-line management. Runs AFTER user rules so that user
//     deletions get post-processed (blank lines collapsed) and
//     so that user-added punctuation gets normalized.
//
// The internal ordering inside ApplyExtra (rationale):
//
//  1. Per-character normalization (full-width → half-width, Kangxi
//     radicals → normal). Doing this first lets later byte-level
//     passes see a uniform ASCII space and uniformly-encoded CJK text.
//  2. Citation-mark deletion (parens / brackets with digit content) is
//     done before any punctuation conversion, so that we don't
//     accidentally reformat the brackets/parens of the citation marks
//     themselves.
//  3. Whitespace normalization — strict newline dedup → newline→space.
//     CollapseSpaces already ran in PreClean; per-line trim and
//     zero-width strip happened there too. Here we only finish the
//     cross-line whitespace shape.
//  4. CJK / digit spacing — runs on the now-clean text, so "delete
//     CJK spaces" happens before "add space after punctuation" and
//     we never accidentally re-collapse what we just added.
//  5. Punctuation conversion (Chinese ↔ English) and Chinese-typography
//     normalization follow the whitespace pass; they touch
//     punctuation marks directly, so we let them have the final say
//     on spacing around them.
//  6. Simplified ↔ Traditional character conversion runs last among
//     the string transforms, since the typography step may have
//     introduced ASCII chars that should remain ASCII after conversion.
//  7. Blank-line management (RemoveEmptyLines / CollapseBlankLines) is
//     the very last step, so it acts on the final shape produced by
//     every prior phase plus any blank lines left behind by user rules.
//
// NormalizeLineEndings / LineEnding are NOT acted on here — processor.go
// does the final \n ↔ CRLF conversion based on these toggles. The two
// entry points only ever see LF-normalized input.
package cleanrules

import (
	"strings"

	"github.com/siongui/gojianfan"

	"textcleaner/internal/model"
)

// arbitrateMutex 防御性地把同一语义轴上的多个 bool 收敛成一个。
//
// 三对互斥项（前端已用 radio 强制，但 settings.json 可能被手编/迁移）：
//   1. 标点方向：只允许 e2c XOR c2e（两个都开 → 留 c2e，即列表中后者）
//   2. 简繁转换：只允许 s2t XOR t2s（两个都开 → 留 t2s，即列表中后者）
//   3. 空白行：remove > collapse > collapseMax（最激进的保留）
//
// 通过指针写入以避免复制整个 opt 大结构体。
func arbitrateMutex(opt *model.BasicCleanOptions) {
	// (1) 标点方向
	if opt.PunctEnglishToChinese && opt.PunctChineseToEnglish {
		// 留 PunctChineseToEnglish（按字段出现顺序，后者更靠后）。
		opt.PunctEnglishToChinese = false
	}
	// (2) 简繁转换
	if opt.SimplifiedToTraditional && opt.TraditionalToSimplified {
		// 留 TraditionalToSimplified（按字段出现顺序，后者更靠后）。
		opt.SimplifiedToTraditional = false
	}
	// (3) 空白行：priority remove > collapse > collapseMax
	if opt.RemoveEmptyLines {
		// RemoveEmptyLines 是最强清理，子选项全让位。
		opt.CollapseNewlines = false
		opt.CollapseBlankLines = false
	} else if opt.CollapseNewlines {
		// CollapseNewlines 已经把多换行压到 1，再跑 CollapseBlankLines 没意义。
		opt.CollapseBlankLines = false
	}
}

// ApplyExtra runs every "post-clean" rule in BasicCleanOptions and returns
// the transformed text. It is meant to be called AFTER the user-supplied
// find/replace rules so that:
//
//   - deletions by the user (e.g. removing an ad block) leave behind
//     blank lines that get post-collapsed by RemoveEmptyLines /
//     CollapseBlankLines;
//   - whitespace, punctuation, and S↔T rewrites run on the final text
//     shape rather than getting re-undone by user rules.
//
// Pre-clean (BOM, per-line trim, zero-width strip, collapse-spaces) is
// exposed as a separate PreClean function — it must run BEFORE user
// rules so that user-supplied find strings match reliably regardless
// of source formatting (whitespace / BOM / zero-width differences).
//
// Mutex arbitration (defensive): the frontend already enforces
// mutual exclusion for three pairs of toggles via radio buttons:
//
//   - 标点方向：punctEnglishToChinese XOR punctChineseToEnglish
//   - 简繁转换：simplifiedToTraditional XOR traditionalToSimplified
//   - 空白行：collapseNewlines / removeEmptyLines / collapseBlankLines
//     （三者同时只有一个为真）
//
// If a legacy or hand-edited settings.json has multiple of these set,
// this function picks a single winner so the output is deterministic:
//   - punctuation / script: keep the one that's flagged (whichever
//     was last in the struct order wins on draw); the other is zeroed
//     out. Running both back-to-back would mostly cancel out.
//   - blank-line: priority is removeEmptyLines > collapseNewlines >
//     collapseBlankLines (matches how a user would visually rank
//     "most aggressive" first).
//
// Pure function: does not mutate opt.
func ApplyExtra(text string, opt model.BasicCleanOptions) string {
	// 防御性仲裁：万一 settings.json 中多个互斥项都开着，统一收成一个。
	// 优先保留“列表顺序靠后”的语义其实没什么意义；这里就按明显的取舍来。
	arbitrateMutex(&opt)

	// Always start with the pre-clean subset (BOM, per-line trim, zero-width,
	// collapse-spaces). It is idempotent, so it's safe to call here even when
	// processor.go already invoked PreClean before the user rules.
	text = PreClean(text, opt)

	// Phase 1: per-character normalization (run AFTER user rules — full-width
	// input that the user wants to match against still has full-width chars).
	if opt.FullWidthToHalfWidth {
		text = fullToHalf(text)
	}
	if opt.ReplaceKangxiRadicals {
		text = replaceKangxiRadicals(text)
	}

	// Phase 2: citation-mark deletion (before any punctuation rewrite).
	if opt.RemoveCiteParen {
		text = removeCitationParen(text)
	}
	if opt.RemoveCiteBracket {
		text = removeCitationBracket(text)
	}

	// Phase 3: whitespace / newlines (dependency-ordered).
	// CollapseSpaces already ran in PreClean — re-running here would be
	// idempotent but wasteful.
	if opt.CollapseNewlines {
		text = collapseNewlines(text)
	}
	if opt.NewlineToSpace {
		text = newlinesToSpace(text)
	}

	// Phase 4: CJK / digit spacing.
	if opt.RemoveSpaceBetweenCJK {
		text = removeSpacesBetweenNonASCII(text)
	}
	if opt.RemoveSpaceAtDecimal {
		text = removeSpaceAtDecimal(text)
	}
	if opt.RemoveSpaceAtColon {
		text = removeSpaceAtColon(text)
	}
	if opt.SpaceBetweenLetterAndDigit {
		text = spaceBetweenLetterAndDigit(text)
	}
	if opt.SpaceAfterPunctuation {
		text = spaceAfterPunctuation(text)
	}

	// Phase 5: punctuation conversion + Chinese typography.
	if opt.PunctChineseToEnglish {
		text = punctChineseToEnglish(text)
	}
	if opt.PunctEnglishToChinese {
		text = punctEnglishToChinese(text)
	}
	if opt.NormalizeChineseTypography {
		text = normalizeChineseTypography(text)
	}

	// Phase 6: S↔T script conversion.
	if opt.SimplifiedToTraditional {
		text = gojianfan.S2T(text)
	}
	if opt.TraditionalToSimplified {
		text = gojianfan.T2S(text)
	}

	// Phase 7: blank-line management (last, so it acts on the final shape
	// produced by all prior phases — including any blank lines that user
	// replacements created by deleting content).
	if opt.RemoveEmptyLines {
		text = removeEmptyLines(text)
	} else if opt.CollapseBlankLines {
		max := opt.MaxBlankLines
		if max < 0 {
			max = 0
		}
		text = collapseBlankLines(text, max)
	}
	return text
}

// PreClean runs the "pre-clean" subset of BasicCleanOptions BEFORE the
// user's find/replace rules. These are line-level normalizations that
// make user-supplied find strings match reliably regardless of the
// source file's formatting quirks (BOM, indentation, full-width spaces,
// zero-width garbage from copy-paste).
//
// The subset covers:
//   - RemoveUTF8BOM       (file-start only)
//   - TrimLeadingWhitespace / TrimTrailingWhitespace  (per line)
//   - RemoveZeroWidthChars (per line, after trim)
//   - CollapseSpaces       (per line, after trim)
//
// Pure function: does not mutate opt.
func PreClean(text string, opt model.BasicCleanOptions) string {
	if opt.RemoveUTF8BOM {
		text = strings.TrimPrefix(text, "\ufeff")
	}
	if !opt.TrimLeadingWhitespace && !opt.TrimTrailingWhitespace &&
		!opt.RemoveZeroWidthChars && !opt.CollapseSpaces {
		return text
	}
	text = perLineTrimAndZeroWidth(
		text,
		opt.TrimLeadingWhitespace,
		opt.TrimTrailingWhitespace,
		opt.RemoveZeroWidthChars,
	)
	if opt.CollapseSpaces {
		text = collapseSpaces(text)
	}
	return text
}

// ---------- character normalization ----------

// fullToHalf maps full-width ASCII punctuation / letters / digits to their
// half-width counterparts. Full-width space (U+3000) becomes ASCII space.
// Implementation note: U+FF01..U+FF5E offset 0xFEE0 = ASCII range 0x21..0x7E.
func fullToHalf(s string) string {
	if !strings.ContainsAny(s, "\u3000\uff01\uff02\uff03\uff04\uff05\uff06\uff07\uff08\uff09\uff0a\uff0b\uff0c\uff0d\uff0e\uff0f\uff10\uff11\uff12\uff13\uff14\uff15\uff16\uff17\uff18\uff19\uff1a\uff1b\uff1c\uff1d\uff1e\uff1f\uff20\uff21\uff22\uff23\uff24\uff25\uff26\uff27\uff28\uff29\uff2a\uff2b\uff2c\uff2d\uff2e\uff2f\uff30\uff31\uff32\uff33\uff34\uff35\uff36\uff37\uff38\uff39\uff3a\uff3b\uff3c\uff3d\uff3e\uff3f\uff40\uff41\uff42\uff43\uff44\uff45\uff46\uff47\uff48\uff49\uff4a\uff4b\uff4c\uff4d\uff4e\uff4f\uff50\uff51\uff52\uff53\uff54\uff55\uff56\uff57\uff58\uff59\uff5a\uff5b\uff5c\uff5d\uff5e") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\u3000':
			b.WriteByte(' ')
		case r >= '\uff01' && r <= '\uff5e':
			b.WriteRune(r - 0xFEE0)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// replaceKangxiRadicals maps the Kangxi radical block (U+2F00..U+2FD5) to
// their modern CJK Unified Ideograph equivalents. We only ship entries that
// have an unambiguous modern equivalent — see rules_data.go for the table.
func replaceKangxiRadicals(s string) string {
	if !strings.ContainsRune(s, '\u2F00') && !strings.ContainsRune(s, '\u2FD5') {
		// Fast-path: most text contains no Kangxi radicals, so avoid building
		// a fresh string. The check covers both endpoints of the block; if
		// either appears, fall through to the per-rune scan.
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if mapped, ok := kangxiRadicalTable[r]; ok {
			b.WriteRune(mapped)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ---------- citation-mark deletion ----------

// removeCitationParen deletes parenthesized digit sequences like "(1)",
// "(2,3)", "(4-7)", including an optional leading and trailing horizontal
// whitespace, and the half-width / full-width paren variants.
func removeCitationParen(s string) string {
	return citeParenRE.ReplaceAllString(s, "")
}

// removeCitationBracket does the same for square / lenticular brackets:
// "[1]", "【2,3】".
func removeCitationBracket(s string) string {
	return citeBracketRE.ReplaceAllString(s, "")
}

// ---------- whitespace normalization ----------

// collapseNewlines reduces runs of two or more newlines (CR/LF/CRLF) to a
// single \n. This is stricter than CollapseBlankLines (which preserves a
// configurable max-blank count): this rule is "always exactly one".
func collapseNewlines(s string) string {
	if !strings.Contains(s, "\n") {
		return s
	}
	// Normalize line endings first so we can match just \n.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// Collapse runs of 2+ newlines into one.
	var b strings.Builder
	b.Grow(len(s))
	prevNL := false
	for _, r := range s {
		if r == '\n' {
			if prevNL {
				continue
			}
			prevNL = true
			b.WriteRune(r)
			continue
		}
		prevNL = false
		b.WriteRune(r)
	}
	return b.String()
}

// newlinesToSpace replaces every newline (CR/LF/CRLF) with a single ASCII
// space. Useful for flattening paragraph text into one long line.
func newlinesToSpace(s string) string {
	if !strings.ContainsAny(s, "\n\r") {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.ReplaceAll(s, "\n", " ")
}

// collapseSpaces reduces runs of 2+ ASCII spaces to one. Does NOT touch
// other whitespace (tabs etc.) — those are handled by per-line trim.
func collapseSpaces(s string) string {
	if !strings.Contains(s, "  ") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if r == ' ' {
			if prevSpace {
				continue
			}
			prevSpace = true
			b.WriteRune(r)
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

// perLineTrimAndZeroWidth is the merged successor of preClean's per-line
// loop: it splits the LF-normalized text by \n, applies the three optional
// line-level operations (trim leading space/tab, trim trailing space/tab,
// strip zero-width characters), and rejoins. We do not touch any other
// whitespace (e.g. U+3000 full-width space) here — that's the job of
// fullToHalf (Phase 1).
func perLineTrimAndZeroWidth(text string, trimLeft, trimRight, stripZeroWidth bool) string {
	if !trimLeft && !trimRight && !stripZeroWidth {
		return text
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if trimLeft {
			line = strings.TrimLeft(line, " \t")
		}
		if trimRight {
			line = strings.TrimRight(line, " \t")
		}
		if stripZeroWidth {
			line = stripZeroWidthRunes(line)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// stripZeroWidthRunes removes the common zero-width / BOM characters.
func stripZeroWidthRunes(s string) string {
	if !strings.ContainsAny(s, "\u200b\u200c\u200d\ufeff") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\u200b', '\u200c', '\u200d', '\ufeff':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// removeEmptyLines drops every line whose contents are entirely whitespace.
// Equivalent to "all consecutive blank-line runs of any size → 0 kept".
func removeEmptyLines(text string) string {
	if !strings.Contains(text, "\n") {
		// Single-line (or empty) input: only treat it as blank if it has no
		// non-whitespace content at all.
		if strings.TrimSpace(text) == "" {
			return ""
		}
		return text
	}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// collapseBlankLines collapses runs of blank (all-whitespace) lines down to
// at most max kept lines between non-blank content. max=0 effectively
// removes all blank lines; max<0 is clamped to 0.
//
// Edge case: if the entire file is blank lines, there's no non-blank
// content to "be between", so we fall back to: keep one blank line when
// max>0, otherwise return empty. This avoids the surprising behaviour of
// wiping a file down to "" just because the user wanted to limit blank
// runs to one.
func collapseBlankLines(text string, max int) string {
	if max < 0 {
		max = 0
	}
	lines := strings.Split(text, "\n")

	nonBlank := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonBlank++
		}
	}
	if nonBlank == 0 {
		// Entire file is blank lines.
		if max > 0 {
			return "\n" // one preserved blank line
		}
		return ""
	}

	out := make([]string, 0, len(lines))
	blankRun := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if blankRun < max {
				out = append(out, line)
				blankRun++
			}
			continue
		}
		blankRun = 0
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// removeSpacesBetweenNonASCII drops ASCII spaces that sit between two
// non-ASCII (high-bit) code points — typically CJK characters or full-width
// punctuation. This is the typical "中文 之间 去 空格" cleanup.
func removeSpacesBetweenNonASCII(s string) string {
	if !strings.ContainsRune(s, ' ') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == ' ' && i > 0 && i+1 < len(runes) &&
			runes[i-1] >= 0x80 && runes[i+1] >= 0x80 {
			continue
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

// removeSpaceAtDecimal deletes spaces between a digit, a dot, and a digit:
// "3 . 14" → "3.14".
func removeSpaceAtDecimal(s string) string {
	return decimalSpaceRE.ReplaceAllString(s, "$1.$2")
}

// removeSpaceAtColon deletes spaces between a digit, a colon, and a digit:
// "12 : 30" → "12:30".
func removeSpaceAtColon(s string) string {
	return colonSpaceRE.ReplaceAllString(s, "$1:$2")
}

// spaceBetweenLetterAndDigit inserts an ASCII space at every letter/digit
// boundary: "abc123" → "abc 123", "iPhone15" → "iPhone 15".
// Caveat: version strings like "v2.0" become "v 2.0" — by design.
//
// Implementation note: Go's regexp alternation leaves the unmatched branch's
// capture groups empty, so a single combined pattern (letter|then|digit OR
// digit|then|letter) with "$1 $2" replacement would silently drop one
// character. We do two separate passes instead — clearer and correct.
func spaceBetweenLetterAndDigit(s string) string {
	s = letterDigitRE1.ReplaceAllString(s, "$1 $2")
	s = letterDigitRE2.ReplaceAllString(s, "$1 $2")
	return s
}

// spaceAfterPunctuation inserts a single ASCII space after a Latin
// punctuation mark when followed by a letter (Latin or CJK). Decimal
// numbers like "3.14" are left alone because the right-hand side is a
// digit, not a letter.
func spaceAfterPunctuation(s string) string {
	return punctSpaceRE.ReplaceAllString(s, "$1 $2")
}

// ---------- punctuation conversion ----------

// punctChineseToEnglish maps common full-width Chinese punctuation to their
// half-width / Latin equivalents. This is a 1-to-1 character substitution —
// context-free quotes are approximated (we map both shapes to a single
// straight ASCII variant rather than tracking open/close state).
func punctChineseToEnglish(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if mapped, ok := punctToEnglish[r]; ok {
			b.WriteRune(mapped)
			continue
		}
		b.WriteRune(r)
	}
	// Multi-rune replacements must run as string-level passes because Go
	// map values are per-rune. The Chinese ellipsis "……" becomes ASCII "...".
	s = b.String()
	s = strings.ReplaceAll(s, "\u2026", "...")
	return s
}

// punctEnglishToChinese maps common half-width punctuation to their full-
// width Chinese equivalents. Straight quotes are mapped to a single shape
// (the curly "right" variant U+201D / U+2019); for proper open/close pairing
// users should run a contextual typography normalizer on top.
//
// Period '.' and colon ':' are converted *only* when they are NOT between
// two digits — otherwise "3.14" would become "3。14" and "12:30" would
// become "12：30", both wrong. We protect decimal / time patterns with
// placeholder runes before the per-rune pass, then restore them after.
func punctEnglishToChinese(s string) string {
	// Pre-collapse ASCII "..." to the Chinese two-character ellipsis "……"
	// (GB/T 15834-2011 style). If we let each '.' turn into '。' we'd get
	// "。。。" instead.
	s = strings.ReplaceAll(s, "...", "\u2026\u2026")

	// Protect decimal / time-style "digit . digit" and "digit : digit"
	// sequences by replacing the punctuation with a placeholder that the
	// next pass won't touch. We use U+0001 / U+0002 because they cannot
	// appear in normal text and we own them for the duration of this call.
	s = asciiDecimalRE.ReplaceAllString(s, "$1"+"\x01"+"$2")
	s = asciiTimeRE.ReplaceAllString(s, "$1"+"\x02"+"$2")

	// Now do the safe per-rune substitution. Period and colon are in the
	// map now; the placeholder protection above means they only convert
	// outside decimal/time context.
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if mapped, ok := punctToChinese[r]; ok {
			b.WriteRune(mapped)
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	out = strings.ReplaceAll(out, "\x01", ".")
	out = strings.ReplaceAll(out, "\x02", ":")
	return out
}

// ---------- Chinese typography normalization ----------
//
// Implements the core rules from sspai.com/post/37815 ("中文文案排版指北"):
//   - one space between Chinese and English / digits
//   - degree / percent signs attach directly to the preceding number
// This is a minimal but useful subset; more aggressive rules (e.g. ellipse
// conversion, "x × y" math spacing) can be added by the user as regex rules.
func normalizeChineseTypography(s string) string {
	s = cjkToAsciiLetterRE.ReplaceAllString(s, "$1 $2")
	s = asciiLetterToCJKRE.ReplaceAllString(s, "$1 $2")
	s = cjkToAsciiDigitRE.ReplaceAllString(s, "$1 $2")
	s = asciiDigitToCJKRE.ReplaceAllString(s, "$1 $2")
	// Degree and percent signs stick to the preceding digit.
	s = degreeRE.ReplaceAllString(s, "$1°")
	s = percentRE.ReplaceAllString(s, "$1%")
	return s
}
