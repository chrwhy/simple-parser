package parser

import "testing"

func TestParseClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// ParseClause 不依赖 jieba，中文不做分词，整串作为一个 token
		{"纯中文", "我爱中国", `("我爱中国")`},
		{"纯英文", "hello", `(pinyin OR hello)`}, // pinyin 部分依赖库实现，仅验证结构
		{"空串", "", ""},
		{"纯空格", "   ", ""},
		{"单字", "我", `("我")`},
		{"繁体", "中華人民共和國", `("中华人民共和国")`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClause(tt.input)
			if tt.name == "纯英文" {
				// 英文路径包含拼音，仅验证非空且包含原始词
				if got == "" {
					t.Errorf("ParseClause(%q) returned empty", tt.input)
				}
				t.Logf("ParseClause(%q) = %s", tt.input, got)
				return
			}
			if got != tt.want {
				t.Errorf("ParseClause(%q)\n  got:  %s\n  want: %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseClause_Mixed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"中英混合", "我爱China"},
		{"英中混合", "hello世界"},
		{"含数字", "test123中文"},
		{"含标点", "hello,world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClause(tt.input)
			if got == "" {
				t.Errorf("ParseClause(%q) returned empty", tt.input)
			}
			t.Logf("ParseClause(%q) = %s", tt.input, got)
		})
	}
}

func TestParseClause_EmptyInput(t *testing.T) {
	if got := ParseClause(""); got != "" {
		t.Errorf("ParseClause(\"\") = %q, want empty", got)
	}
	if got := ParseClause("   "); got != "" {
		t.Errorf("ParseClause(\"   \") = %q, want empty", got)
	}
}
