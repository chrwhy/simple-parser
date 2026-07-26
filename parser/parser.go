package parser

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/yanyiwu/gojieba"
)

//go:embed data/synonym.txt
var defaultSynonymData string

// ParseOption 解析选项。
type ParseOption func(*parseConfig)

type parseConfig struct {
	useSynonym bool
}

// EnableSynonym 启用同义词扩展。
// 需要在 New() 时通过 WithSynonym 或 WithSynonymFile 加载词典。
func EnableSynonym() ParseOption {
	return func(c *parseConfig) {
		c.useSynonym = true
	}
}

// Parser 封装 jieba 分词器生命周期，用于将查询解析为 FTS5 MATCH 子句。
type Parser struct {
	jieba   *gojieba.Jieba
	synonym *SynonymDict
}

// WithSynonym 使用内置同义词词典。
func WithSynonym() func(*Parser) {
	return func(p *Parser) {
		dict, err := parseSynonymDict(strings.NewReader(defaultSynonymData))
		if err != nil {
			return
		}
		p.synonym = dict
	}
}

// WithSynonymFile 从指定文件路径加载同义词词典。
func WithSynonymFile(path string) func(*Parser) {
	return func(p *Parser) {
		dict, err := LoadSynonymDict(path)
		if err != nil {
			return
		}
		p.synonym = dict
	}
}

// New 创建 Parser 并加载 jieba 词典。
// 可选参数：
//   - WithSynonym(): 使用内置同义词词典
//   - WithSynonymFile(path): 从自定义文件加载同义词词典
func New(opts ...func(*Parser)) *Parser {
	p := &Parser{jieba: gojieba.NewJieba()}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Close 释放 jieba 资源。
func (p *Parser) Close() {
	if p.jieba != nil {
		p.jieba.Free()
		p.jieba = nil
	}
}

var (
	defaultParser *Parser
	defaultMu     sync.Mutex
)

// InitJieba 初始化包级默认 Parser，供 ParseJiebaClause 使用。
// 可选参数与 New() 相同，例如 WithSynonym()、WithSynonymFile(path)。
// 未传入任何参数时，默认使用内置同义词词典。
func InitJieba(opts ...func(*Parser)) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultParser != nil {
		defaultParser.Close()
	}
	if len(opts) == 0 {
		opts = []func(*Parser){WithSynonym()}
	}
	defaultParser = New(opts...)
}

// FreeJieba 释放包级默认 Parser。
func FreeJieba() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultParser != nil {
		defaultParser.Close()
		defaultParser = nil
	}
}

// ParseJiebaClause 使用 jieba 分词将用户输入转换为 FTS5 MATCH 条件。
// 此函数持有锁直到操作完成，防止与 FreeJieba 产生竞态。
func ParseJiebaClause(query string, opts ...ParseOption) string {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultParser == nil {
		defaultParser = New()
	}
	return defaultParser.ParseJiebaClause(query, opts...)
}

// ParseJiebaClause 将 query 解析为 FTS5 MATCH 子句。
// 通过 opts 可选启用同义词扩展（需在 New() 时加载词典）。
func (p *Parser) ParseJiebaClause(query string, opts ...ParseOption) string {
	if !isValidQuery(query) {
		return ""
	}
	cfg := &parseConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	var dict *SynonymDict
	if cfg.useSynonym {
		dict = p.synonym
	}
	return parseJiebaClause(p.jieba, query, dict)
}

// LookupSynonymPhrase 在默认 parser 的同义词词典中整词查找短语的同义词。
func LookupSynonymPhrase(phrase string) []string {
	defaultMu.Lock()
	p := defaultParser
	defaultMu.Unlock()
	if p == nil {
		return nil
	}
	return p.synonym.LookupPhrase(phrase)
}

// BuildPhraseSynonymClause 构建整词匹配 + 同义词扩展的 FTS5 MATCH 子句。
// 返回格式如：`"红烧肉" OR "东坡肉"`，所有词作为整词匹配。
// 若无同义词则返回 `"红烧肉"`。
// 调用方自行决定何时使用整词匹配（如短查询场景）。
func BuildPhraseSynonymClause(query string) string {
	phrase := strings.TrimSpace(query)
	syns := LookupSynonymPhrase(phrase)
	clause := `"` + escapeQuote(phrase) + `"`
	for _, syn := range syns {
		clause += ` OR "` + escapeQuote(syn) + `"`
	}
	return clause
}

// isValidQuery 检查查询是否有效：非空、去空格后非空、不超过最大长度。
func isValidQuery(query string) bool {
	if len(query) == 0 {
		return false
	}
	if len(query) > MaxQueryLen*4 { // 快速字节级检查，避免极端长输入触发 rune 遍历
		return false
	}
	trimmed := strings.TrimSpace(query)
	return len(trimmed) > 0 && len([]rune(trimmed)) <= MaxQueryLen
}
