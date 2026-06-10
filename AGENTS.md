# AGENTS.md — simple-parser

> 面向 AI 的项目协作约束文档。描述本 SDK 的设计边界、MATCH 子句生成规则与禁止行为。

---

## 项目概述

将用户搜索词解析为 **SQLite FTS5 MATCH 子句** 的 Go SDK，配合 [chrwhy/simple](https://github.com/chrwhy/simple) 分词扩展使用。

**包名**：`parser`（`import "github.com/chrwhy/simple-parser"`）

**技术栈**：Go 1.23 + gojieba（中文分词）+ open-pinyin（英文拼音候选）

**下游消费者**：
- [chrwhy/simple-search](https://github.com/chrwhy/simple-search) — FTS5 全文检索演示
- [cooking_agent](https://github.com/chrwhy/cooking_agent) — 菜谱 Agent 搜索

**相关文档**：
- [docs/SPEC.md](docs/SPEC.md) — MATCH 子句生成行为规格与示例（改解析规则时须同步更新）

---

## 目录结构

```
simple-parser/
├── doc.go              # 包级文档
├── parser.go           # Parser 生命周期、InitJieba/FreeJieba
├── normalize.go        # ToSimplified 简繁归一化
├── token.go            # token 分类与工具函数
├── jieba.go            # ParseJiebaClause 实现
├── pinyin.go           # ParsePinyinClause
├── clause.go           # ParseClause（轻量路径）
├── data/t2s.txt        # 繁简字典（与 chrwhy/simple 同源）
├── jieba_test.go       # 主测试套件
├── normalize_test.go   # 简繁归一化测试
├── docs/SPEC.md
├── AGENTS.md
└── README.md
```

## 文件职责

| 文件 | 职责 |
|------|------|
| `parser.go` | `Parser` 类型、`New`/`Close`、包级 `InitJieba`/`FreeJieba` |
| `jieba.go` | `ParseJiebaClause` — Jieba 分词 + token 分类 → MATCH 子句 |
| `pinyin.go` | `ParsePinyinClause` — 英文 → 拼音 FTS5 子句 |
| `clause.go` | `ParseClause` — 轻量路径，空格/中英边界切分 |
| `token.go` | token 分类、`splitMixed`、`escapeQuote`、`IsAllEn` |
| `normalize.go` | `ToSimplified` — 繁简归一化，字典 `data/t2s.txt` |
| `data/t2s.txt` | 繁→简逐字映射表（与 chrwhy/simple 保持一致） |
| `doc.go` | 包级文档注释 |
| `jieba_test.go` | 主测试套件（表驱动 + 日志输出） |

---

## 三条解析路径

### 1. ParseJiebaClause（推荐，需 Jieba）

解析前先调用 `ToSimplified`，再 Jieba 分词并按 token 类型生成子句，token 之间用 ` AND ` 连接。

| token 类型 | 规则 | 示例输入片段 | 示例子句 |
|------------|------|--------------|----------|
| 中文 | 双引号包裹，内部引号转义 | `中国` | `"中国"` |
| 英文 | 转小写；有拼音候选时 `(pinyin OR token)`，否则 `("token")` | `Jay` | `(j+a+y OR jay)` |
| 数字 | 双引号包裹精确匹配 | `123` | `"123"` |
| 混合（中英数） | 按字符类型拆分后递归解析 | `test123` | 拆为英文 + 数字两部分 |
| 标点/其他 | 双引号包裹 | `@#$` | `"@#$"` |

### 2. ParsePinyinClause（无需 Jieba）

将纯英文输入转为拼音 FTS5 子句：
- 音节用 `+` 连接（如 `j+a+y`）
- 多组拼音候选用 ` OR ` 连接
- 子拼音（`dict.SUB_PINYIN`）在非末尾且长度 > 1 时，用 `SubPinyinStopSign`（值为 3）作为分隔符包裹

### 3. ParseClause（轻量，无需 Jieba）

解析前先调用 `ToSimplified`，再按空格切分、按中英边界拆分，逻辑类似 Jieba 路径的英文/非英文分支，但不做 Jieba 分词。

### 4. ToSimplified（简繁归一化）

- 字典：`data/t2s.txt`（`繁:简` 格式，与 chrwhy/simple 同源）
- 算法：逐 rune 查表替换，与 simple 的 `PinYin::get_ts` 一致
- 自动应用于 `ParseJiebaClause` 和 `ParseClause` 入口
- 更新字典时须从 `../simple/contrib/t2s.txt` 同步

---

## 生命周期管理

**推荐：实例化 Parser**

```go
p := parser.New()
defer p.Close()
clause := p.ParseJiebaClause("周杰伦")
```

**兼容旧代码：包级函数**

```go
parser.InitJieba()
defer parser.FreeJieba()
clause := parser.ParseJiebaClause("我爱中国")
```

注意：
- `ParseJiebaClause` 包级函数通过 `defaultParserInstance()` 懒初始化
- `ParsePinyinClause` / `ParseClause` 不依赖 Jieba，可直接调用
- `InitJieba` 会替换已有默认实例；测试中用 `TestMain` 统一初始化/释放

---

## 依赖库

| 类别 | 库 | 说明 |
|------|----|------|
| 中文分词 | `github.com/yanyiwu/gojieba` | CGO，加载词典，`NewJieba`/`Free` 管理生命周期 |
| 拼音解析 | `github.com/chrwhy/open-pinyin` | `parser.Parse` / `ParseInitial`、`dict.SUB_PINYIN`、`util.Concat` |

**编译**：纯 Go 包，无额外 build tag。下游若配合 SQLite FTS5 使用，需 `--tags fts5`。

---

## MATCH 子句格式约定

本库输出的是 FTS5 `MATCH` 表达式的**条件部分**，由调用方嵌入 SQL：

```sql
SELECT * FROM docs WHERE docs MATCH ?
```

格式要点：
- 中文/数字/标点：双引号短语 `"token"`
- 英文有拼音：`(拼音子句 OR 原文小写)`，如 `(j+a+y OR jay)`
- 多 token：` AND ` 连接
- 引号转义：双引号 → `""`，单引号 → `''`（`escapeQuote`）

**不要**在本库输出外层 SQL 引号或 `WHERE` 子句；只负责 MATCH 条件字符串。

---

## 测试

```bash
go test ./...
```

测试策略：
- 中文纯词、数字：严格断言期望输出
- 英文、混合、特殊字符：主要断言非空 + `t.Logf` 记录实际输出

**修改解析规则时**：必须更新对应测试的 `want` 或补充表驱动用例。

---

## 禁止行为

1. **不要**改变 MATCH 子句语法而不更新测试和 README 示例
2. **不要**在 `ParsePinyinClause` 中引入 Jieba 依赖
3. **不要**在输出中拼接 SQL 关键字或参数占位符
4. **不要**移除 `InitJieba`/`FreeJieba` 包级 API
5. **不要**修改 `SubPinyinStopSign` 而不验证与 chrwhy/simple tokenizer 的兼容性
6. **不要**在 `classifyToken` / `splitMixed` 中忽略 Unicode 分类（中文用 `unicode.Han`）
7. **不要**替换 `data/t2s.txt` 为第三方 OpenCC 词典而不与 simple 对齐

---

## 常见改动场景

| 场景 | 涉及文件 | 注意 |
|------|----------|------|
| 调整中文分词行为 | `jieba.go` | Jieba 词典在 gojieba 包内 |
| 调整拼音展开规则 | `pinyin.go` | 依赖 open-pinyin |
| 新增 token 类型 | `token.go` + `jieba.go` | 同步更新 `classifyToken` 和 switch |
| 轻量解析增强 | `clause.go` | 与 Jieba 路径行为尽量一致 |
| 简繁字典更新 | `data/t2s.txt` | 从 simple 同步 |

---

## 本地联调

```go
require github.com/chrwhy/simple-parser v0.0.0
replace github.com/chrwhy/simple-parser => ../simple-parser
```
