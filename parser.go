package parser

import (
	"strings"
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
// 此函数持有锁直到操作完成，防止与 FreeJieba 产生竞态。
func ParseJiebaClause(query string) string {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultParser == nil {
		defaultParser = New()
	}
	return defaultParser.ParseJiebaClause(query)
}

// ParseJiebaClause 将 query 解析为 FTS5 MATCH 子句。
func (p *Parser) ParseJiebaClause(query string) string {
	if !isValidQuery(query) {
		return ""
	}
	return parseJiebaClause(p.jieba, query)
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
