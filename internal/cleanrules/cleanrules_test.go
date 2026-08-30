package cleanrules

import (
	"strings"
	"testing"

	"textcleaner/internal/model"
)

// on is a helper to spell out "this option is on" without repeating the
// struct literal. Tests below set only the fields they care about.
func on(flags ...func(*model.BasicCleanOptions)) model.BasicCleanOptions {
	var o model.BasicCleanOptions
	for _, f := range flags {
		f(&o)
	}
	return o
}

// ============================================================
// Phase 1 — BOM + character-level
// ============================================================

func TestRemoveUTF8BOM(t *testing.T) {
	in := "\ufeffHello\n世界\n"
	want := "Hello\n世界\n"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.RemoveUTF8BOM = true }))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// BOM not at the start should NOT be stripped (only TrimPrefix behaviour).
	in2 := "Hello\ufeffWorld"
	if got := ApplyExtra(in2, on(func(o *model.BasicCleanOptions) { o.RemoveUTF8BOM = true })); got != "Hello\ufeffWorld" {
		t.Fatalf("mid-string BOM should be preserved by RemoveUTF8BOM, got %q", got)
	}
}

func TestFullWidthToHalfWidth_LettersAndDigits(t *testing.T) {
	// U+FF21..U+FF23 = full-width A B C; U+FF11..U+FF13 = full-width 1 2 3.
	in := "\uff21\uff22\uff23\uff11\uff12\uff13"
	want := "ABC123"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.FullWidthToHalfWidth = true }))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFullWidthToHalfWidth_Space(t *testing.T) {
	// U+3000 (ideographic space) → ASCII space
	in := "你\u3000好"
	want := "你 好"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.FullWidthToHalfWidth = true }))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFullWidthToHalfWidth_Disabled(t *testing.T) {
	in := "\uff21\uff22"
	if got := ApplyExtra(in, model.BasicCleanOptions{}); got != in {
		t.Fatalf("disabled option should not transform: got %q, want %q", got, in)
	}
}

func TestReplaceKangxiRadicals(t *testing.T) {
	// U+2F00 is the Kangxi radical for "一" (one); the modern equivalent is U+4E00.
	in := "前\u2f00事"
	want := "前一事"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.ReplaceKangxiRadicals = true }))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// ============================================================
// Phase 2 — citation-mark deletion
// ============================================================

func TestRemoveCitationParen(t *testing.T) {
	cases := []struct{ in, want string }{
		{"句子(1)结尾", "句子结尾"},
		{"句子(2,3)结尾", "句子结尾"},
		{"句子(4-7)结尾", "句子结尾"},
		{"带空格 (1) 测试", "带空格测试"}, // both surrounding spaces are consumed with the citation
		{"无括号1无", "无括号1无"},
		{"全角（1）测试", "全角测试"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.RemoveCiteParen = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("RemoveCiteParen(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRemoveCitationBracket(t *testing.T) {
	cases := []struct{ in, want string }{
		{"参考[1]结尾", "参考结尾"},
		{"参考[2,3]结尾", "参考结尾"},
		{"参考[4-7]结尾", "参考结尾"},
		{"全角【1】测试", "全角测试"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.RemoveCiteBracket = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("RemoveCitationBracket(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ============================================================
// Phase 3 — whitespace / newlines
// ============================================================

func TestCollapseNewlines(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\n\n\nb", "a\nb"},
		{"a\r\n\r\n\r\nb", "a\nb"}, // CRLF normalized to LF first
		{"a\n\nb", "a\nb"},
		{"a\nb", "a\nb"}, // single newline stays
	}
	opt := on(func(o *model.BasicCleanOptions) { o.CollapseNewlines = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("CollapseNewlines(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNewlineToSpace(t *testing.T) {
	if got := ApplyExtra("a\nb\rc\r\nd", on(func(o *model.BasicCleanOptions) { o.NewlineToSpace = true })); got != "a b c d" {
		t.Fatalf("NewlineToSpace: got %q", got)
	}
}

func TestCollapseSpaces(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a    b", "a b"},
		{"a  b  c", "a b c"},
		{"a b", "a b"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.CollapseSpaces = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("CollapseSpaces(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTrimLeadingAndTrailingWhitespace(t *testing.T) {
	in := "  hello \t\n   world   \n\t  done  "
	want := "hello\nworld\ndone"
	opt := on(
		func(o *model.BasicCleanOptions) { o.TrimLeadingWhitespace = true },
		func(o *model.BasicCleanOptions) { o.TrimTrailingWhitespace = true },
	)
	if got := ApplyExtra(in, opt); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRemoveZeroWidthChars(t *testing.T) {
	in := "Hello\u200bWorld\u200c!\u200d\ufeffEnd"
	want := "HelloWorld!End"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.RemoveZeroWidthChars = true }))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRemoveEmptyLines(t *testing.T) {
	// Cleanrules only removes blank lines; preservation of a trailing
	// newline (if the original file had one) is processor.ProcessText's
	// job. So the expectations here are cleanrules-only outputs.
	cases := []struct{ in, want string }{
		{"a\n\n\nb\n   \nc\n", "a\nb\nc"},
		{"\n\n\n", ""},
		{"a", "a"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.RemoveEmptyLines = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("RemoveEmptyLines(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCollapseBlankLines(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"a\n\n\nb", 1, "a\n\nb"},   // collapse 3 blank lines down to 1 kept between content
		{"a\n\n\n\n\nb", 2, "a\n\n\nb"}, // collapse to max 2 blank lines
		{"a\n\n\nb", 0, "a\nb"},   // max=0 means remove all blank lines
		{"\n\n\n", 1, "\n"},      // file that is just blank lines
	}
	for _, c := range cases {
		opt := on(func(o *model.BasicCleanOptions) {
			o.CollapseBlankLines = true
			o.MaxBlankLines = c.max
		})
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("CollapseBlankLines(%q, max=%d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}

// RemoveEmptyLines takes precedence over CollapseBlankLines (matches the
// previous pre/post-clean behaviour).
func TestRemoveEmptyLines_OverridesCollapseBlankLines(t *testing.T) {
	in := "a\n\n\nb"
	opt := on(
		func(o *model.BasicCleanOptions) { o.RemoveEmptyLines = true },
		func(o *model.BasicCleanOptions) { o.CollapseBlankLines = true },
		func(o *model.BasicCleanOptions) { o.MaxBlankLines = 2 },
	)
	if got := ApplyExtra(in, opt); got != "a\nb" {
		t.Fatalf("RemoveEmptyLines should win over CollapseBlankLines: got %q", got)
	}
}

// ============================================================
// Phase 4 — CJK / digit spacing
// ============================================================

func TestRemoveSpacesBetweenNonASCII(t *testing.T) {
	cases := []struct{ in, want string }{
		{"你 好", "你好"},
		{"中 国 人", "中国人"},
		{"你 good 朋友", "你 good 朋友"}, // ASCII boundary untouched
		{"a b c", "a b c"},              // pure ASCII untouched
	}
	opt := on(func(o *model.BasicCleanOptions) { o.RemoveSpaceBetweenCJK = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("RemoveSpacesBetweenNonASCII(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRemoveSpaceAtDecimal(t *testing.T) {
	if got := ApplyExtra("3 . 14", on(func(o *model.BasicCleanOptions) { o.RemoveSpaceAtDecimal = true })); got != "3.14" {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveSpaceAtColon(t *testing.T) {
	if got := ApplyExtra("12 : 30", on(func(o *model.BasicCleanOptions) { o.RemoveSpaceAtColon = true })); got != "12:30" {
		t.Fatalf("got %q", got)
	}
}

func TestSpaceBetweenLetterAndDigit(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc123", "abc 123"},
		{"123abc", "123 abc"},
		{"iPhone15", "iPhone 15"},
		{"plain text", "plain text"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.SpaceBetweenLetterAndDigit = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("SpaceBetweenLetterAndDigit(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSpaceAfterPunctuation(t *testing.T) {
	cases := []struct{ in, want string }{
		{"hello,world", "hello, world"},
		{"foo;bar", "foo; bar"},
		{"price:3.14", "price:3.14"}, // '.' is followed by a digit, not a letter
		{"你好,世界", "你好, 世界"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.SpaceAfterPunctuation = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("SpaceAfterPunctuation(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ============================================================
// Phase 5 — punctuation + Chinese typography
// ============================================================

func TestPunctChineseToEnglish(t *testing.T) {
	cases := []struct{ in, want string }{
		{"你好，世界。", "你好,世界."},
		{"你好！再见？", "你好!再见?"},
		{"（测试）", "(测试)"},
		{"【引用】", "[引用]"},
		{"等\u2026", "等..."}, // ellipsis → three dots
	}
	opt := on(func(o *model.BasicCleanOptions) { o.PunctChineseToEnglish = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("PunctChineseToEnglish(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPunctEnglishToChinese(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello, World.", "Hello， World。"},
		{"foo;bar", "foo；bar"},
		{"Wait!Really?", "Wait！Really？"},
		{"(test)", "（test）"},
		{"[ref]", "【ref】"},
		{"3.14", "3.14"},        // decimal point stays
		{"12:30", "12:30"},      // time separator stays
		{"see...", "see……"},     // three dots → Chinese ellipsis
	}
	opt := on(func(o *model.BasicCleanOptions) { o.PunctEnglishToChinese = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("PunctEnglishToChinese(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeChineseTypography_BasicSpacing(t *testing.T) {
	// Naive "always add space at digit↔CJK boundary" rule — matches the
	// simplest reading of sspai post 37815. More nuanced rules (no space
	// between year number and 年, etc.) would require word-boundary
	// detection and are out of scope for the "基础功能" feature set.
	cases := []struct{ in, want string }{
		{"今天Apple发布", "今天 Apple 发布"},
		{"我买了3本书", "我买了 3 本书"},
		{"这里有5只猫", "这里有 5 只猫"},
		{"2025年第1季度", "2025 年第 1 季度"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.NormalizeChineseTypography = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("NormalizeChineseTypography(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeChineseTypography_DegreePercent(t *testing.T) {
	cases := []struct{ in, want string }{
		{"今天 30 °C", "今天 30°C"},
		{"完成 50 %", "完成 50%"},
	}
	opt := on(func(o *model.BasicCleanOptions) { o.NormalizeChineseTypography = true })
	for _, c := range cases {
		if got := ApplyExtra(c.in, opt); got != c.want {
			t.Errorf("NormalizeChineseTypography(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ============================================================
// Phase 6 — Simplified ↔ Traditional
// ============================================================

func TestSimplifiedToTraditional(t *testing.T) {
	if got := ApplyExtra("简体中文", on(func(o *model.BasicCleanOptions) { o.SimplifiedToTraditional = true })); got != "簡體中文" {
		t.Fatalf("S2T: got %q, want 簡體中文", got)
	}
}

func TestTraditionalToSimplified(t *testing.T) {
	if got := ApplyExtra("繁體中文", on(func(o *model.BasicCleanOptions) { o.TraditionalToSimplified = true })); got != "繁体中文" {
		t.Fatalf("T2S: got %q, want 繁体中文", got)
	}
}

func TestSimplifiedTraditionalRoundtrip(t *testing.T) {
	in := "简体中文转换测试 123 abc"
	s2t := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.SimplifiedToTraditional = true }))
	t2s := ApplyExtra(s2t, on(func(o *model.BasicCleanOptions) { o.TraditionalToSimplified = true }))
	if t2s != in {
		t.Fatalf("roundtrip: %q → %q → %q", in, s2t, t2s)
	}
}

// ============================================================
// Combined / disabled / no-mutation
// ============================================================

func TestAllDisabledNoOp(t *testing.T) {
	in := "原文不动(1)【2】简体中文  。hello, world 3.14"
	if got := ApplyExtra(in, model.BasicCleanOptions{}); got != in {
		t.Fatalf("disabled option should leave text untouched: got %q, want %q", got, in)
	}
}

func TestCombinedRulesOrder(t *testing.T) {
	// End-to-end: collapse newlines, then spaces, then strip citation
	// marks, then convert English punctuation to Chinese. Make sure
	// the pipeline produces the expected combined result.
	in := "你好,世界.\n\n(1) 价格 3.14 元"
	opt := on(
		func(o *model.BasicCleanOptions) { o.CollapseNewlines = true },
		func(o *model.BasicCleanOptions) { o.CollapseSpaces = true },
		func(o *model.BasicCleanOptions) { o.RemoveCiteParen = true },
		func(o *model.BasicCleanOptions) { o.PunctEnglishToChinese = true },
	)
	want := "你好，世界。\n价格 3.14 元"
	got := ApplyExtra(in, opt)
	if got != want {
		t.Fatalf("combined:\n got %q\nwant %q", got, want)
	}
}

func TestApplyExtra_DoesNotMutateOpts(t *testing.T) {
	in := "abc"
	original := model.BasicCleanOptions{
		FullWidthToHalfWidth:       true,
		SpaceBetweenLetterAndDigit: true,
	}
	_ = ApplyExtra(in, original)
	if !original.FullWidthToHalfWidth || !original.SpaceBetweenLetterAndDigit {
		t.Fatal("ApplyExtra must not mutate its options argument")
	}
}

// ============================================================
// 互斥项仲裁（前端用 radio 强制，后端做防御性的"只能跑一个"）
// ============================================================

// 三组两两/三三互斥的 bool 同时为真时，ApplyExtra 只跑其中一个；
// 输出对每组应有稳定的"赢家"，不能因为两个都跑而出错/抵消。

func TestApplyExtra_Mutex_Punct_BothE2CAndC2E_OnlyOneWins(t *testing.T) {
	// 文本里既有中文逗号，又有英文逗号——跑 c2e 后全是 ASCII；跑 e2c 后全是全角。
	// 仲裁应只跑其中一个；这里只验证不报错、输出是有意义的。
	in := "你好,世界"
	both := ApplyExtra(in, on(func(o *model.BasicCleanOptions) {
		o.PunctEnglishToChinese = true
		o.PunctChineseToEnglish = true
	}))
	// 期望：输出 100% 是其中一种形状，不会出现"逗号还是逗号但视觉上没变"的抵消。
	if strings.Contains(both, ",") && strings.Contains(both, "，") {
		t.Fatalf("mutex arbitration failed: mixed punctuation shapes: %q", both)
	}
	// 既不是 ASCII 也不是 full-width，说明出 bug 了
	if !strings.Contains(both, ",") && !strings.Contains(both, "，") {
		t.Fatalf("mutex arbitration produced unexpected output: %q", both)
	}
}

func TestApplyExtra_Mutex_Script_BothS2TAndT2S_OnlyOneWins(t *testing.T) {
	// S→T 跑完再 T→S 会基本回到原文（不动）；但如果只跑一个就会有变化。
	// 验证仲裁后输出不等于 neither 跑（要么简要么繁，不能完全没动）。
	in := "简体中文"
	both := ApplyExtra(in, on(func(o *model.BasicCleanOptions) {
		o.SimplifiedToTraditional = true
		o.TraditionalToSimplified = true
	}))
	onlyS2T := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.SimplifiedToTraditional = true }))
	onlyT2S := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.TraditionalToSimplified = true }))
	if both != onlyS2T && both != onlyT2S {
		t.Fatalf("mutex arbitration should pick exactly one of S2T/T2S; got %q vs %q / %q", both, onlyS2T, onlyT2S)
	}
}

func TestApplyExtra_Mutex_BlankLine_RemoveWinsOverCollapse(t *testing.T) {
	in := "a\n\n\n\nb\n\n  \nc"
	all := ApplyExtra(in, on(func(o *model.BasicCleanOptions) {
		o.RemoveEmptyLines = true
		o.CollapseNewlines = true
		o.CollapseBlankLines = true
	}))
	want := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.RemoveEmptyLines = true }))
	if all != want {
		t.Fatalf("remove+collapse+collapseMax should equal remove alone: got %q, want %q", all, want)
	}
}

func TestApplyExtra_Mutex_BlankLine_CollapseWinsOverCollapseBlank(t *testing.T) {
	in := "a\n\n\n\nb"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) {
		o.CollapseNewlines = true
		o.CollapseBlankLines = true
	}))
	want := ApplyExtra(in, on(func(o *model.BasicCleanOptions) { o.CollapseNewlines = true }))
	if got != want {
		t.Fatalf("collapse + collapseMax should equal collapse alone: got %q, want %q", got, want)
	}
}

func TestApplyExtra_Mutex_AllClear_NoRejection(t *testing.T) {
	// 没有任何互斥项为真时不受影响：不能因为仲裁逻辑而误清选项。
	in := "你好, world  。  简"
	got := ApplyExtra(in, on(func(o *model.BasicCleanOptions) {
		o.PunctEnglishToChinese = false
		o.PunctChineseToEnglish = false
		o.SimplifiedToTraditional = false
		o.TraditionalToSimplified = false
		o.CollapseNewlines = false
		o.RemoveEmptyLines = false
		o.CollapseBlankLines = false
	}))
	if got != in {
		t.Fatalf("all mutex options off should leave text unchanged: got %q, want %q", got, in)
	}
}

// Sanity check that we don't accidentally call the (un-exported) punctuation
// string helpers without `strings` — would have failed at compile time but
// the import is used elsewhere via strings.ReplaceAll, so guard anyway.
var _ = strings.Contains
