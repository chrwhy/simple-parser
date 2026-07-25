package parser

import (
	"strings"

	"github.com/yanyiwu/gojieba"
)

func parseJiebaClause(j *gojieba.Jieba, query string, synonym *SynonymDict) string {
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
			part := `"` + word + `"`
			// 同义词扩展
			if synonym != nil {
				synonyms := synonym.Lookup(word)
				for _, syn := range synonyms {
					syn = escapeQuote(syn)
					part += ` OR "` + syn + `"`
				}
			}
			// 有同义词时需要括号，确保 AND 优先级正确
			if synonym != nil && len(synonym.Lookup(word)) > 0 {
				part = "(" + part + ")"
			}
			parts = append(parts, part)
		case tokenAlpha:
			lower := strings.ToLower(word)
			pinyinClause := ParsePinyinClause(lower)
			var part string
			if len(pinyinClause) > 0 {
				part = "(" + pinyinClause + " OR " + lower
			} else {
				part = `("` + lower + `"`
			}
			// 同义词扩展
			if synonym != nil {
				synonyms := synonym.Lookup(lower)
				for _, syn := range synonyms {
					syn = strings.ToLower(syn)
					part += " OR " + syn
				}
			}
			part += ")"
			parts = append(parts, part)
		case tokenDigit:
			parts = append(parts, `"`+word+`"`)
		case tokenMixed:
			for _, st := range splitMixed(word) {
				sub := parseJiebaClause(j, st, synonym)
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
