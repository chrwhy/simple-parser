package parser

import (
	"strings"
)

// ParseClause 按空格与中英边界切分 query，并生成 FTS5 MATCH 子句（不依赖 jieba）。
func ParseClause(query string) string {
	if !isValidQuery(query) {
		return ""
	}
	query = ToSimplified(query)
	spaceTokens := strings.Split(query, " ")

	regroupedTokens := make([]string, 0)
	for _, token := range spaceTokens {
		enCnTokens := splitCnEnToken(token)
		if enCnTokens == nil {
			regroupedTokens = append(regroupedTokens, token)
		} else {
			regroupedTokens = append(regroupedTokens, enCnTokens...)
		}
	}

	var b strings.Builder
	for _, token := range regroupedTokens {
		if IsAllEn(token) {
			pinyinClause := ParsePinyinClause(token)
			if b.Len() > 0 {
				b.WriteString(" AND ")
			}
			if len(pinyinClause) > 0 {
				b.WriteByte('(')
				b.WriteString(pinyinClause)
				b.WriteString(" OR ")
				b.WriteString(token)
				b.WriteByte(')')
			} else {
				b.WriteString(`("`)
				b.WriteString(token)
				b.WriteString(`")`)
			}
		} else {
			token = escapeQuote(token)
			if b.Len() > 0 {
				b.WriteString(" AND ")
			}
			b.WriteString(`("`)
			b.WriteString(token)
			b.WriteString(`")`)
		}
	}

	return b.String()
}

func splitCnEnToken(input string) []string {
	return splitCnEn(input)
}
