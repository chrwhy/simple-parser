package parser

import (
	"strings"
	"unicode"
)

const (
	// MaxQueryLen 是单次查询的最大字符数（rune），超出则返回空字符串。
	MaxQueryLen = 1024
)

type tokenType int

const (
	tokenChinese tokenType = iota
	tokenAlpha
	tokenDigit
	tokenMixed
	tokenPunct
)

func classifyToken(s string) tokenType {
	hasChinese := false
	hasAlpha := false
	hasDigit := false

	for _, r := range s {
		switch {
		case unicode.Is(unicode.Han, r):
			hasChinese = true
		case unicode.IsLetter(r):
			hasAlpha = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	mixed := 0
	if hasChinese {
		mixed++
	}
	if hasAlpha {
		mixed++
	}
	if hasDigit {
		mixed++
	}
	if mixed > 1 {
		return tokenMixed
	}
	if hasChinese {
		return tokenChinese
	}
	if hasAlpha {
		return tokenAlpha
	}
	if hasDigit {
		return tokenDigit
	}
	return tokenPunct
}

func splitMixed(s string) []string {
	var result []string
	var buf strings.Builder
	var lastType rune

	for _, r := range s {
		var t rune
		switch {
		case unicode.Is(unicode.Han, r):
			t = 'C'
		case unicode.IsLetter(r):
			t = 'E'
		case unicode.IsDigit(r):
			t = 'D'
		default:
			t = 'O'
		}

		if lastType != 0 && t != lastType {
			if buf.Len() > 0 {
				result = append(result, buf.String())
				buf.Reset()
			}
		}
		buf.WriteRune(r)
		lastType = t
	}
	if buf.Len() > 0 {
		result = append(result, buf.String())
	}
	return result
}

func escapeQuote(s string) string {
	s = strings.ReplaceAll(s, `"`, `""`)
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

// IsAllEn 判断字符串是否全部为英文字母。
// 纯 ASCII 检查，直接遍历字节避免 []rune 分配。
func IsAllEn(query string) bool {
	if len(query) == 0 {
		return false
	}
	for i := 0; i < len(query); i++ {
		c := query[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

// splitCnEn 按中英文边界切分字符串。
// 数字和标点归入相邻的 token，不单独产生切分点。
func splitCnEn(input string) []string {
	var result []string
	var buf strings.Builder
	var lastClass rune // 'C'=中文, 'E'=英文, 'O'=其他(数字/标点)

	for _, r := range input {
		var class rune
		switch {
		case unicode.Is(unicode.Han, r):
			class = 'C'
		case unicode.IsLetter(r):
			class = 'E'
		default:
			class = 'O'
		}

		// 仅在中文↔英文边界切分；数字/标点跟随当前 token。
		if lastClass != 0 && isCnEnBoundary(lastClass, class) {
			if buf.Len() > 0 {
				result = append(result, buf.String())
				buf.Reset()
			}
		}
		buf.WriteRune(r)
		// 只有中文和英文会改变 lastClass，数字/标点延续上一个。
		if class != 'O' {
			lastClass = class
		}
	}
	if buf.Len() > 0 {
		result = append(result, buf.String())
	}

	// 未产生切分则返回 nil，让调用方回退到原始 token。
	if len(result) <= 1 {
		return nil
	}
	return result
}

// isCnEnBoundary 判断两个字符类之间是否构成中英文边界。
func isCnEnBoundary(a, b rune) bool {
	return (a == 'C' && b == 'E') || (a == 'E' && b == 'C')
}
