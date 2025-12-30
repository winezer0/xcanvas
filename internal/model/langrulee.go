package model

// DynamicCategory 动态分类规则
// - Category: 分类结果
// - FilePatterns: 关联文件模式（满足任一即可）
// - Dependencies: 关联依赖包（满足任一即可）
type DynamicCategory struct {
	Category     string   `json:"category"`
	FilePatterns []string `json:"file_patterns"`
	Dependencies []string `json:"dependencies"`
}

// Language 统一语言模型，整合语言特征和分类规则
// - Name: 语言名称（如 "Go", "JavaScript"）
// - LineComments: 行注释标记
// - MultiLine: 多行注释标记
// - Extensions: 文件扩展名
// - Filenames: 特定文件名
// - Category: 默认分类（frontend/backend/desktop/other）
// - Dynamic: 动态分类规则列表
type Language struct {
	Name         string            `json:"name"`
	LineComments []string          `json:"line_comments"`
	MultiLine    [][]string        `json:"multi_line"`
	Extensions   []string          `json:"extensions"`
	Filenames    []string          `json:"filenames"`
	Category     string            `json:"category"`
	Dynamic      []DynamicCategory `json:"dynamic"`
}
