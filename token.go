package parser

import (
	"strings"
	"unicode"
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
func IsAllEn(query string) bool {
	for _, r := range []rune(query) {
		if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') {
			return false
		}
	}
	return true
}

func splitCnEn(input string) []string {
	var result []string
	var current string
	var currentType rune

	for _, r := range input {
		var charType rune
		if unicode.Is(unicode.Han, r) {
			charType = 'C'
		} else if unicode.IsLetter(r) {
			charType = 'E'
		} else {
			return nil
		}

		if currentType == 0 {
			currentType = charType
		}

		if charType != currentType {
			if current != "" {
				result = append(result, current)
			}
			current = string(r)
			currentType = charType
		} else {
			current += string(r)
		}
	}

	if current != "" {
		result = append(result, current)
	}
	return result
}
