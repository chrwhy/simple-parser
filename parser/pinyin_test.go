package parser

import "testing"

func TestParsePinyinClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"单音节", "ma"},
		{"双音节", "zhongguo"},
		{"三音节", "zhoujielun"},
		{"单字母", "a"},
		{"空串", ""},
		{"纯空格", "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePinyinClause(tt.input)
			if tt.input == "" || tt.input == "   " {
				if got != "" {
					t.Errorf("ParsePinyinClause(%q) = %q, want empty", tt.input, got)
				}
				return
			}
			if got == "" {
				t.Errorf("ParsePinyinClause(%q) returned empty", tt.input)
			}
			t.Logf("ParsePinyinClause(%q) = %s", tt.input, got)
		})
	}
}

func TestParsePinyinClause_NonEmpty(t *testing.T) {
	// 验证常见英文名能产生有意义的拼音子句
	names := []string{"jay", "chou", "test", "hello", "world"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			got := ParsePinyinClause(name)
			if got == "" {
				t.Errorf("ParsePinyinClause(%q) returned empty", name)
			}
			t.Logf("ParsePinyinClause(%q) = %s", name, got)
		})
	}
}
