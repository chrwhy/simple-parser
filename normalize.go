package parser

import (
	"bufio"
	_ "embed"
	"strings"
	"sync"
)

// t2s.txt 与 chrwhy/simple/contrib/t2s.txt 保持一致（繁:简 逐字映射）。
//
//go:embed data/t2s.txt
var t2sData string

var (
	t2sMap  map[rune]rune
	t2sOnce sync.Once
)

func initT2SMap() {
	t2sMap = make(map[rune]rune)
	scanner := bufio.NewScanner(strings.NewReader(t2sData))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		idx := strings.IndexByte(line, ':')
		if idx <= 0 {
			continue
		}
		t := []rune(line[:idx])
		s := []rune(line[idx+1:])
		if len(t) == 1 && len(s) == 1 {
			t2sMap[t[0]] = s[0]
		}
	}
}

// ToSimplified 将字符串中的繁体中文转为简体，非映射字符保持不变。
// 算法与 chrwhy/simple 的 PinYin::get_ts 一致：逐 rune 查 t2s 字典。
func ToSimplified(s string) string {
	t2sOnce.Do(initT2SMap)
	runes := []rune(s)
	changed := false
	for i, r := range runes {
		if simp, ok := t2sMap[r]; ok {
			runes[i] = simp
			changed = true
		}
	}
	if !changed {
		return s
	}
	return string(runes)
}
