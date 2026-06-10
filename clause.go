package parser

import (
	"strings"
)

// ParseClause 按空格与中英边界切分 query，并生成 FTS5 MATCH 子句（不依赖 jieba）。
func ParseClause(query string) string {
	query = ToSimplified(query)
	clause := ""
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

	for _, token := range regroupedTokens {
		if IsAllEn(token) {
			pinyinClause := ParsePinyinClause(token)
			partialSql := ""
			if len(clause) > 0 {
				partialSql = " AND "
			}
			if len(pinyinClause) > 0 {
				partialSql = partialSql + `(` + pinyinClause + " OR " + token + `)`
			} else {
				partialSql = partialSql + `("` + token + `")`
			}
			clause = clause + partialSql
		} else {
			token = escapeQuote(token)
			sql := `("` + token + `")`
			if len(clause) > 0 {
				clause = clause + " AND " + sql
			} else {
				clause = sql
			}
		}
	}

	return clause
}

func splitCnEnToken(input string) []string {
	return splitCnEn(input)
}
