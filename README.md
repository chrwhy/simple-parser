# simple-parser

将用户搜索词解析为 **SQLite FTS5 MATCH 子句** 的 Go SDK，配合 [chrwhy/simple](https://github.com/chrwhy/simple) 分词扩展使用。

## 文档

| 文档 | 说明 |
|------|------|
| [AGENTS.md](AGENTS.md) | AI 协作文档：设计边界、禁止行为、改动指引 |
| [docs/SPEC.md](docs/SPEC.md) | MATCH 子句生成行为规格与示例 |

## 功能

- **ParseJiebaClause**：简繁归一化 + Jieba 分词 + 中英数分类，生成 AND 连接的 MATCH 子句
- **ParsePinyinClause**：英文转拼音候选（基于 [open-pinyin](https://github.com/chrwhy/open-pinyin)）
- **ParseClause**：简繁归一化 + 不依赖 Jieba 的空格/中英切分解析（轻量路径）
- **ToSimplified**：繁体转简体（字典与 [chrwhy/simple](https://github.com/chrwhy/simple) 的 `contrib/t2s.txt` 同源）

## 安装

```bash
go get github.com/chrwhy/simple-parser
```

本地联调（与 simple-search 同目录时）：

```go
// go.mod
require github.com/chrwhy/simple-parser v0.0.0
replace github.com/chrwhy/simple-parser => ../simple-parser
```

## 用法

### 推荐：实例化 Parser

```go
import "github.com/chrwhy/simple-parser"

p := parser.New()
defer p.Close()

clause := p.ParseJiebaClause("周杰伦 Jay Chou")
// `"周杰伦" AND (j+a+y OR jay) AND (ch+ou OR chou)`

// 繁体输入自动转简体
clause = p.ParseJiebaClause("中華人民共和國")
// `"中华人民共和国"`
```

### 包级函数（兼容旧代码）

```go
parser.InitJieba()
defer parser.FreeJieba()

clause := parser.ParseJiebaClause("我爱China")
pinyin := parser.ParsePinyinClause("china") // 无需 jieba
```

## 开发与测试

```bash
go test ./...
```

## 依赖

- `github.com/yanyiwu/gojieba` — 中文分词
- `github.com/chrwhy/open-pinyin` — 英文拼音 query 扩展
