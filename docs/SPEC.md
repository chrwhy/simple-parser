# MATCH 子句生成规格

本文档描述 `simple-parser` 的输出行为。实现以本规格与 `jieba_test.go` 为准；规格变更须同步更新测试。

## 范围

- **输入**：用户搜索词（UTF-8 字符串）
- **输出**：SQLite FTS5 `MATCH` 表达式的**条件部分**（不含 `WHERE`、不含 SQL 外层引号）
- **不负责**：建表、`tokenize='simple'` 配置、SQL 参数绑定

典型嵌入方式：

```sql
SELECT rowid, highlight(docs, 1, '<b>', '</b>') FROM docs WHERE docs MATCH ?
```

---

## API 概览

| 函数 | 依赖 Jieba | 用途 |
|------|------------|------|
| `ToSimplified` | 否 | 繁体 → 简体（`data/t2s.txt` 逐字映射） |
| `ParseJiebaClause` | 是 | 简繁归一化 + 中文分词 + 全类型 token 处理 |
| `ParsePinyinClause` | 否 | 纯英文 → 拼音 FTS5 子句 |
| `ParseClause` | 否 | 简繁归一化 + 空格/中英边界切分 |

---

## 简繁归一化（ToSimplified）

### 字典来源

与 [chrwhy/simple](https://github.com/chrwhy/simple) 的 `contrib/t2s.txt` 保持一致（本仓库位于 `data/t2s.txt`），格式为 `繁:简`（逐字一行）。

### 处理规则

1. 逐 rune 查表，命中则替换为简体
2. 未命中字符（英文、数字、已是简体等）保持不变
3. `ParseJiebaClause` / `ParseClause` 在分词前自动调用

### 示例

| 输入 | 输出 |
|------|------|
| `中華人民共和國` | `中华人民共和国` |
| `我愛中國` | `我爱中国` |
| `hello` | `hello` |

### 严格断言

| 输入 | `ParseJiebaClause` 期望输出 |
|------|----------------------------|
| `中華人民共和國` | `"中华人民共和国"` |

---

## ParseJiebaClause

### 处理流程

1. 调用 `ToSimplified(query)` 繁简归一化
2. 使用 Jieba `Cut(query, true)` 分词（HMM 新词识别开启）
3. 对每个词去除首尾空白，空词跳过
4. 调用 `classifyToken` 分类
5. 按类型生成子句片段
6. 非空片段用 ` AND ` 连接

### Token 分类（classifyToken）

按 Unicode 统计字符类型：

| 条件 | 类型 |
|------|------|
| 同时含中文、英文、数字中的两种及以上 | `tokenMixed` |
| 仅中文（`unicode.Han`） | `tokenChinese` |
| 仅英文字母 | `tokenAlpha` |
| 仅数字 | `tokenDigit` |
| 其他（标点、符号等） | `tokenPunct` |

### 各类型输出规则

#### 中文（tokenChinese）

- 对内容执行 `escapeQuote`（`"` → `""`，`'` → `''`）
- 输出：`"` + 内容 + `"`

#### 英文（tokenAlpha）

- 转小写 `lower`
- 调用 `ParsePinyinClause(lower)`
- 若拼音子句非空：`(拼音子句 OR lower)`
- 否则：`("lower")`

#### 数字（tokenDigit）

- 输出：`"` + 原文 + `"`（不做 escape，因纯数字无引号）

#### 混合（tokenMixed）

- 调用 `splitMixed` 按中文/英文/数字/其他边界拆分
- 对每个子串递归 `parseJiebaClause`
- 非空子结果加入 parts（最终仍由顶层 ` AND ` 连接）

#### 标点/其他（tokenPunct）

- 同中文：`escapeQuote` 后双引号包裹

### 边界输入

| 输入 | 输出 |
|------|------|
| `""` | `""` |
| 仅空白 | `""` |
| 首尾空白 | 分词前由 Jieba 处理，一般仍能产出有效子句 |

### 严格断言示例（测试用例）

| 输入 | 期望输出 |
|------|----------|
| `我爱中国` | `"我" AND "爱" AND "中国"` |
| `周杰伦` | `"周杰伦"` |
| `中华人民共和国` | `"中华人民共和国"` |
| `13825638962` | `"13825638962"` |
| `123` | `"123"` |
| `8` | `"8"` |

### 典型输出示例

| 输入 | 输出（参考） |
|------|--------------|
| `周杰伦 Jay Chou` | `"周杰伦" AND (j+a+y OR jay) AND (ch+ou OR chou)` |
| `我爱China` | `"我" AND "爱" AND (c+h+i+n+a OR china)` |

英文、混合、特殊字符等用例以 `go test -v ./...` 日志为准。

---

## ParsePinyinClause

### 处理流程

1. `pinyin.Parse(input)` 获取拼音分组
2. `pinyin.ParseInitial(input)` 获取声母分组；非空时追加到分组列表
3. 对每个分组内音节：
   - 若在 `dict.SUB_PINYIN` 中，且非末音节且长度 > 1：包裹为 `"音节" + SubPinyinStopSign`（`SubPinyinStopSign = 3`）
4. 组内音节用 `+` 连接（`util.Concat`）
5. 多组之间用 ` OR ` 连接

### 示例

| 输入 | 输出形态 |
|------|----------|
| `jay` | `j+a+y`（可能含额外 OR 声母候选） |
| `china` | 多音节 `+` 连接，可能含 OR 分支 |

具体输出依赖 `open-pinyin` 词典与 `SUB_PINYIN` 配置。

### 约束

- 不调用 Jieba
- 输入通常为小写英文（Jieba 路径会在调用前转小写）

---

## ParseClause

### 处理流程

1. 调用 `ToSimplified(query)` 繁简归一化
2. 按空格 `Split` 得到 tokens
2. 对每个 token，尝试 `splitCnEnToken`（仅含中文与英文字母时按边界拆分）
3. 对每个最终 token：
   - **全英文**（`IsAllEn`）：逻辑同 Jieba 路径的英文分支（拼音 OR 原文）
   - **非全英文**：`escapeQuote`（仅处理 `"` 和 `'`）后 `("token")`，多 token 用 ` AND ` 连接

### 与 ParseJiebaClause 的差异

| 维度 | ParseJiebaClause | ParseClause |
|------|------------------|-------------|
| 中文切分 | Jieba 词典分词 | 不按词切，整段或空格/中英边界 |
| 混合 token | `splitMixed` 递归 | 仅空格与中英边界 |
| 数字/标点 | 完整 classifyToken | 非全英文整段引号包裹 |

轻量路径适合无 Jieba 开销场景；语义可能与 Jieba 路径不一致，选用时需明确。

---

## 引号转义（escapeQuote）

用于中文、标点及 `ParseClause` 非英文分支：

| 字符 | 替换为 |
|------|--------|
| `"` | `""` |
| `'` | `''` |

---

## splitMixed 字符类型

混合 token 按以下类型边界拆分：

| 类型码 | 含义 |
|--------|------|
| `C` | 中文（`unicode.Han`） |
| `E` | 英文字母 |
| `D` | 数字 |
| `O` | 其他 |

相邻不同类型字符构成拆分点。

---

## 版本与兼容性

- Go 1.23+
- `SubPinyinStopSign = 3` 须与 [chrwhy/simple](https://github.com/chrwhy/simple) tokenizer 子拼音分隔约定一致
- 修改输出格式视为**破坏性变更**，需通知 `simple-search`、`cooking_agent` 等下游
