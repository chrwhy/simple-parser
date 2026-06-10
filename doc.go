// Package parser 将用户搜索词解析为 SQLite FTS5 MATCH 子句。
//
// 典型用法：
//
//	p := parser.New()
//	defer p.Close()
//	clause := p.ParseJiebaClause("周杰伦 Jay Chou")
//
// 也可使用包级函数（需先调用 InitJieba / 结束时 FreeJieba）：
//
//	parser.InitJieba()
//	defer parser.FreeJieba()
//	clause := parser.ParseJiebaClause("周杰伦")
package parser
