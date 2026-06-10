package parser

import (
	"sync"

	"github.com/yanyiwu/gojieba"
)

// Parser 封装 jieba 分词器生命周期，用于将查询解析为 FTS5 MATCH 子句。
type Parser struct {
	jieba *gojieba.Jieba
}

// New 创建 Parser 并加载 jieba 词典。
func New() *Parser {
	return &Parser{jieba: gojieba.NewJieba()}
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
func InitJieba() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultParser != nil {
		defaultParser.Close()
	}
	defaultParser = New()
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
func ParseJiebaClause(query string) string {
	return defaultParserInstance().ParseJiebaClause(query)
}

// ParseJiebaClause 将 query 解析为 FTS5 MATCH 子句。
func (p *Parser) ParseJiebaClause(query string) string {
	return parseJiebaClause(p.jieba, query)
}

func defaultParserInstance() *Parser {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultParser == nil {
		defaultParser = New()
	}
	return defaultParser
}
