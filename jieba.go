package parser

import (
	"strings"

	"github.com/yanyiwu/gojieba"
)

func parseJiebaClause(j *gojieba.Jieba, query string) string {
	query = ToSimplified(query)
	words := j.Cut(query, true)

	var parts []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		switch classifyToken(word) {
		case tokenChinese:
			word = escapeQuote(word)
			parts = append(parts, `"`+word+`"`)
		case tokenAlpha:
			lower := strings.ToLower(word)
			pinyinClause := ParsePinyinClause(lower)
			if len(pinyinClause) > 0 {
				parts = append(parts, "("+pinyinClause+" OR "+lower+")")
			} else {
				parts = append(parts, `("`+lower+`")`)
			}
		case tokenDigit:
			parts = append(parts, `"`+word+`"`)
		case tokenMixed:
			for _, st := range splitMixed(word) {
				sub := parseJiebaClause(j, st)
				if sub != "" {
					parts = append(parts, sub)
				}
			}
		case tokenPunct:
			word = escapeQuote(word)
			parts = append(parts, `"`+word+`"`)
		}
	}

	return strings.Join(parts, " AND ")
}
