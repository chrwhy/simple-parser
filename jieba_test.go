package parser

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	InitJieba()
	code := m.Run()
	FreeJieba()
	os.Exit(code)
}

func TestParseJiebaClause_ChineseOnly(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"我爱中国", `"我" AND "爱" AND "中国"`},
		{"周杰伦", `"周杰伦"`},
		{"中华人民共和国", `"中华人民共和国"`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got != tt.want {
				t.Errorf("ParseJiebaClause(%q)\n  got:  %s\n  want: %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseJiebaClause_EnglishOnly(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"hello"},
		{"a"},
		{"Jay Chou"},
		{"test"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParseJiebaClause_Mixed(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"我爱China"},
		{"周杰伦 Jay Chou"},
		{"hello世界"},
		{"test123中文"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParseJiebaClause_SpecialChars(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`he said "hello"`},
		{"it's fine"},
		{"@#$%"},
		{"@English &special"},
		{`"bacon-&and"-eggs%`},
		{"---"},
		{"..."},
		{"hello!world"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParseJiebaClause_EmptyAndSpaces(t *testing.T) {
	got := ParseJiebaClause("")
	if got != "" {
		t.Errorf("empty input: got %q, want empty", got)
	}

	got = ParseJiebaClause("   ")
	if got != "" {
		t.Errorf("spaces only: got %q, want empty", got)
	}
}

func TestParseJiebaClause_Digits(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"纯数字_手机号", "13825638962", `"13825638962"`},
		{"纯数字_短号", "123", `"123"`},
		{"纯数字_单位数", "8", `"8"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got != tt.want {
				t.Errorf("ParseJiebaClause(%q)\n  got:  %s\n  want: %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseJiebaClause_DigitMixed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"数字+中文", "2024年"},
		{"中文+数字", "第3季度"},
		{"英文+数字", "test123"},
		{"数字+英文", "3DMax"},
		{"数字+符号+中文", "100%完成"},
		{"IP地址", "192.168.1.1"},
		{"版本号", "v2.0.1"},
		{"带区号电话", "010-88888888"},
		{"金额", "¥99.9"},
		{"日期", "2024-03-26"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParseJiebaClause_SymbolMixed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"中文@中文", "珠海@中国"},
		{"邮箱", "test@gmail.com"},
		{"文件路径_unix", "/usr/local/bin"},
		{"文件路径_win", "C:\\Users\\test"},
		{"URL", "https://example.com"},
		{"括号包裹", "(测试)"},
		{"中括号", "[重要]通知"},
		{"中文标点", "你好，世界！"},
		{"混合标点", "hello,world;foo:bar"},
		{"连字符短语", "well-known"},
		{"下划线", "hello_world"},
		{"加号", "C++"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParseJiebaClause_ComplexMixed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"中英数符号全混", "订单#2024001:iPhone15 Pro已发货"},
		{"实际搜索_文件名", "报告2024Q1_final.docx"},
		{"实际搜索_地址", "北京市海淀区中关村South大街1号"},
		{"实际搜索_产品", "华为Mate60 Pro+ 256GB"},
		{"实际搜索_日志", "[ERROR] 2024-03-26 连接超时 timeout=30s"},
		{"繁简混合", "中華人民共和國 China"},
		{"重复空格", "hello   world   你好"},
		{"首尾空格", "  周杰伦  "},
		{"单个中文字", "我"},
		{"单个英文字母", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseJiebaClause(tt.input)
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParser_ParseJiebaClause(t *testing.T) {
	p := New()
	defer p.Close()

	got := p.ParseJiebaClause("我爱中国")
	want := `"我" AND "爱" AND "中国"`
	if got != want {
		t.Errorf("Parser.ParseJiebaClause(%q)\n  got:  %s\n  want: %s", "我爱中国", got, want)
	}
}

func TestParseJiebaClause_TraditionalChinese(t *testing.T) {
	got := ParseJiebaClause("中華人民共和國")
	want := `"中华人民共和国"`
	if got != want {
		t.Errorf("ParseJiebaClause(繁体)\n  got:  %s\n  want: %s", got, want)
	}
}

func TestParseJiebaClause_PrintExamples(t *testing.T) {
	examples := []string{
		"我爱中国",
		"周杰伦 Jay Chou",
		"hello世界",
		"北京清华大学",
		"全文检索",
		"test",
		"zhoujielun",
		"13825638962",
		"珠海@中国",
		"test@gmail.com",
		"2024年第3季度",
		"订单#2024001:iPhone15已发货",
		"[ERROR] timeout=30s",
		"/usr/local/bin",
		"华为Mate60 Pro+",
	}
	fmt.Println("\n=== Jieba Parser 分词效果 ===")
	for _, q := range examples {
		result := ParseJiebaClause(q)
		fmt.Printf("  %-30s -> %s\n", q, result)
	}
	fmt.Println("=============================")
}
