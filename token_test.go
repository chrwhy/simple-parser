package parser

import "testing"

func TestClassifyToken(t *testing.T) {
	tests := []struct {
		input string
		want  tokenType
	}{
		{"中国", tokenChinese},
		{"hello", tokenAlpha},
		{"123", tokenDigit},
		{"test123", tokenMixed},
		{"你好world", tokenMixed},
		{"2024年", tokenMixed},
		{"@#$", tokenPunct},
		{"", tokenPunct}, // 空串无字符，走 default
		{"！", tokenPunct},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := classifyToken(tt.input)
			if got != tt.want {
				t.Errorf("classifyToken(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitMixed(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"test123", []string{"test", "123"}},
		{"你好world", []string{"你好", "world"}},
		{"2024年", []string{"2024", "年"}},
		{"abc", []string{"abc"}},
		{"123", []string{"123"}},
		{"你好", []string{"你好"}},
		{"hello世界123", []string{"hello", "世界", "123"}},
		{"a1b", []string{"a", "1", "b"}},
		{"", nil},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitMixed(tt.input)
			if len(got) == 0 && len(tt.want) == 0 {
				return // both nil/empty
			}
			if len(got) != len(tt.want) {
				t.Errorf("splitMixed(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitMixed(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitCnEn(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"纯中文", "你好世界", nil},
		{"纯英文", "hello", nil},
		{"中英混合", "hello世界", []string{"hello", "世界"}},
		{"英中混合", "世界hello", []string{"世界", "hello"}},
		{"中英中", "你hello好", []string{"你", "hello", "好"}},
		{"含数字_不切分", "test123", nil},
		{"含标点_不切分", "hello!", nil},
		{"中英数混合", "hello世界123", []string{"hello", "世界123"}},
		{"英中数混合", "hello123世界", []string{"hello123", "世界"}},
		{"空串", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCnEn(tt.input)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("splitCnEn(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitCnEn(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestEscapeQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{`he"llo`, `he""llo`},
		{"it's", "it''s"},
		{`he"it's`, `he""it''s`},
		{"", ""},
		{"noquotes", "noquotes"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeQuote(tt.input)
			if got != tt.want {
				t.Errorf("escapeQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsAllEn(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", true},
		{"Hello", true},
		{"HELLO", true},
		{"abcXYZ", true},
		{"hello123", false},
		{"hello世界", false},
		{"hello!", false},
		{" hello", false},
		{"", false},
		{"a", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsAllEn(tt.input)
			if got != tt.want {
				t.Errorf("IsAllEn(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsCnEnBoundary(t *testing.T) {
	tests := []struct {
		name string
		a, b rune
		want bool
	}{
		{"中→英", 'C', 'E', true},
		{"英→中", 'E', 'C', true},
		{"中→中", 'C', 'C', false},
		{"英→英", 'E', 'E', false},
		{"中→其他", 'C', 'O', false},
		{"英→其他", 'E', 'O', false},
		{"其他→中", 'O', 'C', false},
		{"其他→英", 'O', 'E', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCnEnBoundary(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("isCnEnBoundary(%c, %c) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsValidQuery(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", true},
		{"中国", true},
		{"", false},
		{"   ", false},
		{" a ", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidQuery(tt.input)
			if got != tt.want {
				t.Errorf("isValidQuery(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}

	// 测试超长输入
	longQuery := make([]rune, MaxQueryLen+1)
	for i := range longQuery {
		longQuery[i] = 'a'
	}
	if isValidQuery(string(longQuery)) {
		t.Errorf("isValidQuery(长度 %d) 应返回 false", MaxQueryLen+1)
	}

	exactQuery := make([]rune, MaxQueryLen)
	for i := range exactQuery {
		exactQuery[i] = 'a'
	}
	if !isValidQuery(string(exactQuery)) {
		t.Errorf("isValidQuery(长度 %d) 应返回 true", MaxQueryLen)
	}
}
