package processor

import (
	"strings"
	"testing"

	"textcleaner/internal/model"
)

func opts() model.ProcessOptions {
	return model.ProcessOptions{
		BasicClean: model.BasicCleanOptions{
			TrimLeadingWhitespace:  true,
			TrimTrailingWhitespace: true,
			MaxBlankLines:          1,
		},
	}
}

func TestTrimLeadingTrailingWhitespace(t *testing.T) {
	in := "   你好   \n世界\t  \n"
	out, res, err := ProcessText(in, opts())
	if err != nil {
		t.Fatal(err)
	}
	if out != "你好\n世界\n" {
		t.Fatalf("trim failed: %q", out)
	}
	if !res.Changed {
		t.Fatal("expected changed=true")
	}
}

func TestCollapseBlankLines(t *testing.T) {
	o := opts()
	o.BasicClean.CollapseBlankLines = true
	o.BasicClean.MaxBlankLines = 1
	in := "第一段\n\n\n\n第二段\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "第一段\n\n第二段\n" {
		t.Fatalf("collapse blank lines failed: %q", out)
	}
}

func TestRemoveEmptyLines(t *testing.T) {
	o := opts()
	o.BasicClean.RemoveEmptyLines = true
	in := "a\n\n\nb\n   \nc\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "a\nb\nc\n" {
		t.Fatalf("remove empty lines failed: %q", out)
	}
}

func TestCollapseSpaces(t *testing.T) {
	o := opts()
	o.BasicClean.CollapseSpaces = true
	in := "Hello     World\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hello World\n" {
		t.Fatalf("collapse spaces failed: %q", out)
	}
}

func TestRemoveZeroWidthChars(t *testing.T) {
	o := opts()
	o.BasicClean.RemoveZeroWidthChars = true
	in := "Hello\u200bWorld\u200c!\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(out, '\u200b') || strings.ContainsRune(out, '\u200c') {
		t.Fatalf("zero width not removed: %q", out)
	}
	if out != "HelloWorld!\n" {
		t.Fatalf("zero width result wrong: %q", out)
	}
}

func TestRemoveUTF8BOM(t *testing.T) {
	o := opts()
	o.BasicClean.RemoveUTF8BOM = true
	in := "\ufeffHello\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(out, '\ufeff') {
		t.Fatalf("BOM not removed: %q", out)
	}
	if out != "Hello\n" {
		t.Fatalf("BOM result wrong: %q", out)
	}
}

// 删除只是“替换为空白内容”的特例：下面用 Replace 中 Replace 为空的规则表达。

func TestDeleteByLineAsEmptyReplace(t *testing.T) {
	o := opts()
	// 逐行删除 = 每条内容对应一条 Replace 为空的规则
	o.Replace = []model.ReplaceRule{
		{Enabled: true, Find: "广告A", Replace: ""},
		{Enabled: true, Find: "广告B", Replace: ""},
	}
	in := "正文\n广告A\n中间\n广告B\n结尾\n"
	out, res, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "广告A") || strings.Contains(out, "广告B") {
		t.Fatalf("delete by line failed: %q", out)
	}
	if res.DeletedMatches != 2 {
		t.Fatalf("expected 2 deleted matches, got %d", res.DeletedMatches)
	}
}

func TestDeleteAsBlockAsEmptyReplace(t *testing.T) {
	o := opts()
	// 整段删除 = 一条 Replace 为空的规则
	o.Replace = []model.ReplaceRule{
		{Enabled: true, Find: "第一段\n第二段", Replace: ""},
	}
	in := "头部\n第一段\n第二段\n尾部\n"
	out, res, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "第一段") {
		t.Fatalf("delete as block failed: %q", out)
	}
	if res.DeletedMatches != 1 {
		t.Fatalf("expected 1 block match, got %d", res.DeletedMatches)
	}
}

func TestReplaceNormal(t *testing.T) {
	o := opts()
	o.Replace = []model.ReplaceRule{
		{Enabled: true, Find: "，，", Replace: "，"},
		{Enabled: true, Find: "錯誤", Replace: "错误"},
	}
	in := "这是，，(一个)錯誤示例\n"
	out, res, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "这是，(一个)错误示例\n" {
		t.Fatalf("normal replace failed: %q", out)
	}
	if res.ReplacedMatches != 2 {
		t.Fatalf("expected 2 replaced matches, got %d", res.ReplacedMatches)
	}
}

func TestReplaceRegex(t *testing.T) {
	o := opts()
	o.Replace = []model.ReplaceRule{
		{Enabled: true, Regex: true, Find: `\s+$`, Replace: ""},
	}
	in := "行尾空格   \n下一行\t\t\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "行尾空格\n下一行\n" {
		t.Fatalf("regex replace failed: %q", out)
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	o := opts()
	o.BasicClean.NormalizeLineEndings = true
	o.BasicClean.LineEnding = model.LineEndingCRLF
	in := "a\nb\nc\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "a\r\nb\r\nc\r\n" {
		t.Fatalf("normalize to CRLF failed: %q", out)
	}
}

func TestKeepCRLFWhenNotNormalizing(t *testing.T) {
	o := opts()
	in := "a\r\nb\r\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	if out != "a\r\nb\r\n" {
		t.Fatalf("keep CRLF failed: %q", out)
	}
}

func TestFullPipelineOrder(t *testing.T) {
	// 模拟：删除广告内容后产生空行，后置整理应合并空行
	o := opts()
	o.BasicClean.CollapseBlankLines = true
	o.BasicClean.MaxBlankLines = 1
	o.Replace = []model.ReplaceRule{
		{Enabled: true, Find: "广告", Replace: ""},
	}
	in := "第一段\n\n广告\n\n第二段\n"
	out, _, err := ProcessText(in, o)
	if err != nil {
		t.Fatal(err)
	}
	want := "第一段\n\n第二段\n"
	if out != want {
		t.Fatalf("full pipeline failed:\n got %q\nwant %q", out, want)
	}
}
