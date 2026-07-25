package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// createTestSynonymFile 创建临时同义词词典文件
func createTestSynonymFile(t *testing.T) string {
	t.Helper()
	content := `# 测试同义词词典
电脑:计算机,PC
手机:移动电话,智能手机
快乐:高兴,开心
`
	dir := t.TempDir()
	path := filepath.Join(dir, "synonym.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("创建测试词典失败: %v", err)
	}
	return path
}

func TestLoadSynonymDict(t *testing.T) {
	path := createTestSynonymFile(t)
	dict, err := LoadSynonymDict(path)
	if err != nil {
		t.Fatalf("LoadSynonymDict() error = %v", err)
	}
	if dict.Size() != 3 {
		t.Errorf("Size() = %d, want 3", dict.Size())
	}
}

func TestSynonymLookup_Chinese(t *testing.T) {
	path := createTestSynonymFile(t)
	dict, err := LoadSynonymDict(path)
	if err != nil {
		t.Fatalf("LoadSynonymDict() error = %v", err)
	}

	tests := []struct {
		word string
		want int // 期望的同义词数量
	}{
		{"电脑", 2},      // 计算机, PC
		{"计算机", 2},    // 电脑, PC（双向查找）
		{"PC", 2},        // 电脑, 计算机（双向查找）
		{"手机", 2},      // 移动电话, 智能手机
		{"不存在的词", 0},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			synonyms := dict.Lookup(tt.word)
			if len(synonyms) != tt.want {
				t.Errorf("Lookup(%q) = %v, want %d items", tt.word, synonyms, tt.want)
			}
		})
	}
}

func TestSynonymLookup_Nil(t *testing.T) {
	var dict *SynonymDict
	synonyms := dict.Lookup("任何词")
	if synonyms != nil {
		t.Errorf("Lookup() on nil dict = %v, want nil", synonyms)
	}
}

func TestParseJiebaClause_WithSynonym_Chinese(t *testing.T) {
	path := createTestSynonymFile(t)
	p := New(WithSynonymFile(path))
	defer p.Close()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "中文同义词",
			input: "电脑",
			want:  `("电脑" OR "计算机" OR "PC")`,
		},
		{
			name:  "中文无同义词",
			input: "中国",
			want:  `"中国"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ParseJiebaClause(tt.input, EnableSynonym())
			if got != tt.want {
				t.Errorf("ParseJiebaClause(%q)\n  got:  %s\n  want: %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseJiebaClause_WithSynonym_Mixed(t *testing.T) {
	path := createTestSynonymFile(t)
	p := New(WithSynonymFile(path))
	defer p.Close()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "中英混合",
			input: "电脑 hello",
		},
		{
			name:  "无同义词混合",
			input: "我爱中国",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ParseJiebaClause(tt.input, EnableSynonym())
			if got == "" {
				t.Errorf("ParseJiebaClause(%q) returned empty", tt.input)
			}
			t.Logf("%s -> %s", tt.input, got)
		})
	}
}

func TestParseJiebaClause_DefaultNoSynonym(t *testing.T) {
	// 默认不启用同义词，即使加载了词典
	path := createTestSynonymFile(t)
	p := New(WithSynonymFile(path))
	defer p.Close()

	got := p.ParseJiebaClause("电脑")
	want := `"电脑"`
	if got != want {
		t.Errorf("ParseJiebaClause(%q)\n  got:  %s\n  want: %s", "电脑", got, want)
	}
}

func TestParseJiebaClause_WithoutSynonym(t *testing.T) {
	// 不使用同义词时，行为应该与原来一致
	p := New()
	defer p.Close()

	tests := []struct {
		input string
		want  string
	}{
		{"我爱中国", `"我" AND "爱" AND "中国"`},
		{"周杰伦", `"周杰伦"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := p.ParseJiebaClause(tt.input)
			if got != tt.want {
				t.Errorf("ParseJiebaClause(%q)\n  got:  %s\n  want: %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseJiebaClause_SynonymFileNotExist(t *testing.T) {
	// 词典文件不存在时，应该不启用同义词，不panic
	p := New(WithSynonymFile("/不存在/path/synonym.txt"))
	defer p.Close()

	got := p.ParseJiebaClause("电脑", EnableSynonym())
	if got == "" {
		t.Error("ParseJiebaClause() returned empty")
	}
	t.Logf("无词典时输出: %s", got)
}

func TestParseJiebaClause_WithSynonym_Embedded(t *testing.T) {
	// 使用内置同义词词典
	p := New(WithSynonym())
	defer p.Close()

	got := p.ParseJiebaClause("电脑", EnableSynonym())
	if got == "" {
		t.Error("ParseJiebaClause() returned empty")
	}
	// 内置词典包含 电脑:计算机,PC
	want := `("电脑" OR "计算机" OR "PC")`
	if got != want {
		t.Errorf("ParseJiebaClause(%q)\n  got:  %s\n  want: %s", "电脑", got, want)
	}
}
