package encoding

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// Decode 自动识别文件编码并解码为 Go 字符串（utf8.String）。
//
// 识别顺序：BOM（UTF-8 / UTF-16 LE / UTF-16 BE）→ 合法 UTF-8 →
// GB18030（含 GBK，简体中文）→ Big5（繁体中文，作为回退试探）。
// 返回：解码后的文本、检测到的编码名、是否带有 BOM。
//
// 编码名取值：utf-8 / utf-8-bom / utf-16le / utf-16be / gb18030 / big5
func Decode(data []byte) (text string, encName string, hadBOM bool) {
	// 1) BOM 检测
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return string(data[3:]), "utf-8-bom", true
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		t, _ := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().String(string(data[2:]))
		return t, "utf-16le", true
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		t, _ := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder().String(string(data[2:]))
		return t, "utf-16be", true
	}

	// 2) 合法 UTF-8
	if utf8.Valid(data) {
		return string(data), "utf-8", false
	}

	// 3) 试探 GB18030（简体，GBK 超集）
	gbkText, _ := simplifiedchinese.GB18030.NewDecoder().String(string(data))
	gbkBad := replacementRatio(gbkText)
	if gbkBad < 0.02 {
		return gbkText, "gb18030", false
	}

	// 4) 试探 Big5（繁体）；若比 GB18030 更“干净”则采用
	big5Text, _ := traditionalchinese.Big5.NewDecoder().String(string(data))
	big5Bad := replacementRatio(big5Text)
	if big5Bad < gbkBad {
		return big5Text, "big5", false
	}

	// 回退：仍以 GB18030 尽力解码
	return gbkText, "gb18030", false
}

// replacementRatio 统计解码文本中 U+FFFD 占比，用于判断编码是否选错。
func replacementRatio(s string) float64 {
	if len(s) == 0 {
		return 1
	}
	n := 0
	for _, r := range s {
		if r == '�' {
			n++
		}
	}
	return float64(n) / float64(len([]rune(s)))
}

// OutputEncoding 指定输出编码策略。
type OutputEncoding string

const (
	// OutKeep 保持输入检测到的编码。
	OutKeep OutputEncoding = "keep"
	// OutUTF8 输出 UTF-8（无 BOM）。
	OutUTF8 OutputEncoding = "utf-8"
	// OutUTF8BOM 输出 UTF-8（带 BOM）。
	OutUTF8BOM OutputEncoding = "utf-8-bom"
)

// Encode 将文本编码为字节。srcEnc 为 Decode 检测到的原始编码名，
// 用于 OutKeep 策略下按原编码重新编码。
func Encode(text string, out OutputEncoding, srcEnc string) []byte {
	switch out {
	case OutUTF8:
		return []byte(text)
	case OutUTF8BOM:
		return append([]byte{0xEF, 0xBB, 0xBF}, []byte(text)...)
	case OutKeep, "":
		return encodeWith(text, srcEnc)
	default:
		return []byte(text)
	}
}

func encodeWith(text string, srcEnc string) []byte {
	var enc encoding.Encoding
	switch srcEnc {
	case "gb18030":
		enc = simplifiedchinese.GB18030
	case "big5":
		enc = traditionalchinese.Big5
	default:
		return []byte(text) // utf-8 / utf-16* 等统一回退为 UTF-8 无 BOM
	}
	b, err := enc.NewEncoder().Bytes([]byte(text))
	if err != nil {
		return []byte(text)
	}
	return b
}

// KeepBOM 仅用于说明：对于 OutKeep 策略，我们默认不重新写入 BOM，
// 以避免重复 BOM。需要保留 BOM 的场景由调用方另行处理。

