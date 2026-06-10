package parser

import (
	"strings"

	"github.com/chrwhy/open-pinyin/dict"
	openpinyin "github.com/chrwhy/open-pinyin/parser"
	"github.com/chrwhy/open-pinyin/util"
)

const (
	// SubPinyinStopSign 子拼音分隔符，须与 chrwhy/simple tokenizer 一致。
	SubPinyinStopSign = 3
)

// ParsePinyinClause 将英文输入解析为 FTS5 拼音 MATCH 子句。
func ParsePinyinClause(input string) string {
	if len(strings.TrimSpace(input)) == 0 {
		return ""
	}
	pinyinGroups := openpinyin.Parse(input)
	pinyinInitial := openpinyin.ParseInitial(input)
	if len(pinyinInitial) > 0 {
		pinyinGroups = append(pinyinGroups, pinyinInitial)
	}

	var b strings.Builder
	for i, pinyinGroup := range pinyinGroups {
		for j := range pinyinGroup {
			if _, ok := dict.SUB_PINYIN[pinyinGroup[j]]; ok {
				if j != len(pinyinGroup)-1 && len(pinyinGroup[j]) > 1 {
					pinyinGroup[j] = "\"" + pinyinGroup[j] + string(rune(SubPinyinStopSign)) + "\""
				}
			}
		}
		b.WriteString(util.Concat(pinyinGroup, "+"))
		if len(pinyinGroups) > 1 && i != len(pinyinGroups)-1 {
			b.WriteString(" OR ")
		}
	}
	return b.String()
}
