package parser

import (
	"strings"

	"github.com/yanyiwu/gojieba"
)

// functionalPOS 虚词词性前缀，分词后过滤。
// c=连词 u=助词 p=介词 e=叹词 y=语气词 o=拟声词 w=标点
// 注意: x=非语素字 不过滤，用户词典添加的词会被标为 x
var functionalPOS = map[string]bool{
	"c": true, "u": true, "p": true, "e": true,
	"y": true, "o": true, "w": true,
}

// CutContentWords 使用 jieba 分词并过滤虚词，返回实词列表。
// 通过 Tag() 获取词性，过滤掉连词、助词、介词等虚词。
// 例: "鱼和豆瓣酱" → ["鱼", "豆瓣酱"]（"和"被过滤，词性 c=连词）
func CutContentWords(query string) []string {
	defaultMu.Lock()
	p := defaultParser
	defaultMu.Unlock()
	if p == nil || p.jieba == nil {
		return nil
	}
	return cutContentWords(p.jieba, query)
}

func cutContentWords(j *gojieba.Jieba, query string) []string {
	query = ToSimplified(query)
	tagged := j.Tag(query)
	var result []string
	for _, t := range tagged {
		// Tag 返回 "词/词性" 格式
		parts := strings.SplitN(t, "/", 2)
		if len(parts) != 2 {
			continue
		}
		word := strings.TrimSpace(parts[0])
		pos := parts[1]
		if word == "" {
			continue
		}
		// 过滤虚词（取词性前缀的第一个字符）
		if functionalPOS[pos[:1]] {
			continue
		}
		result = append(result, word)
	}
	return result
}

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
