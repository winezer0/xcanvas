package camodels

// LangInfo  某一编程语言或标记语言的详细统计数据。
type LangInfo struct {
	Name         string `json:"name"` // 例如: "Java", "YAML"
	Files        int    `json:"files"`
	CodeLines    int    `json:"codeLines"`
	CommentLines int    `json:"commentLines"`
	BlankLines   int    `json:"blankLines"`
}

// LangSummary 保存单一语言的统计结果
type LangSummary struct {
	Name    string
	Code    int64
	Comment int64
	Blank   int64
	Count   int64
}
