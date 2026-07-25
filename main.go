package main

import (
	"fmt"
	"os"

	parser "github.com/chrwhy/simple-parser/parser"
)

func main() {
	dictPath := "data/synonym.txt"

	dict, err := parser.LoadSynonymDict(dictPath)
	if err != nil {
		fmt.Printf("加载词典失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== 同义词词典测试 ===\n")
	fmt.Printf("词典条目数: %d\n\n", dict.Size())

	testCases := []struct {
		word    string
		comment string
	}{
		{"电脑", "中文主词"},
		{"计算机", "中文同义词"},
		{"手机", "中文主词"},
		{"快乐", "中文主词"},
		{"高兴", "中文同义词"},
		{"搜索", "中文主词"},
		{"查找", "中文同义词"},
		{"hello", "英文主词"},
		{"hi", "英文同义词"},
		{"computer", "英文主词"},
		{"happy", "英文主词"},
		{"search", "英文主词"},
		{"不存在的词", "不存在的词"},
	}

	for _, tc := range testCases {
		synonyms := dict.Lookup(tc.word)
		fmt.Printf("[%s] %q\n", tc.comment, tc.word)
		if len(synonyms) == 0 {
			fmt.Printf("  -> (无同义词)\n")
		} else {
			fmt.Printf("  -> %v\n", synonyms)
		}
		fmt.Println()
	}

	fmt.Println("=== Parser 集成测试 ===\n")

	p := parser.New(parser.WithSynonym(dictPath))
	defer p.Close()

	queryTests := []string{
		"计算机",
		"电脑",
		"手机 搜索",
		"快乐学习",
		"hello",
		"happy computer",
	}

	for _, q := range queryTests {
		clause := p.ParseJiebaClause(q)
		fmt.Printf("输入: %q\n", q)
		fmt.Printf("  -> FTS5: %s\n\n", clause)
	}
}
