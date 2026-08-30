package encoding

import (
	"bytes"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
)

func TestDecodeUTF8BOM(t *testing.T) {
	src := []byte{0xEF, 0xBB, 0xBF, 'h', 'i'}
	text, enc, hadBOM := Decode(src)
	if text != "hi" || enc != "utf-8-bom" || !hadBOM {
		t.Fatalf("got text=%q enc=%q bom=%v", text, enc, hadBOM)
	}
}

func TestDecodeUTF8Plain(t *testing.T) {
	text, enc, hadBOM := Decode([]byte("你好world"))
	if text != "你好world" || enc != "utf-8" || hadBOM {
		t.Fatalf("got text=%q enc=%q bom=%v", text, enc, hadBOM)
	}
}

func TestDecodeGBK(t *testing.T) {
	raw, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte("中文测试"))
	if err != nil {
		t.Fatal(err)
	}
	// GBK 字节不应是合法 UTF-8
	if utf8.Valid(raw) {
		t.Fatal("GBK bytes unexpectedly valid UTF-8")
	}
	text, enc, _ := Decode(raw)
	if text != "中文测试" || enc != "gb18030" {
		t.Fatalf("got text=%q enc=%q", text, enc)
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	body, err := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder().Bytes([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	raw := append([]byte{0xFF, 0xFE}, body...)
	text, enc, hadBOM := Decode(raw)
	if text != "hello" || enc != "utf-16le" || !hadBOM {
		t.Fatalf("got text=%q enc=%q bom=%v", text, enc, hadBOM)
	}
}

func TestEncodeRoundTrip(t *testing.T) {
	// keep 策略：按原编码重新编码
	raw, _ := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte("中文"))
	out := Encode("中文", OutKeep, "gb18030")
	if !bytes.Equal(out, raw) {
		t.Fatalf("keep-encode mismatch: %x vs %x", out, raw)
	}
	// utf-8-bom 策略：带 BOM
	out2 := Encode("hi", OutUTF8BOM, "utf-8")
	if len(out2) < 3 || out2[0] != 0xEF || out2[1] != 0xBB || out2[2] != 0xBF {
		t.Fatalf("utf-8-bom missing BOM: %x", out2)
	}
}
