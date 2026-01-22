package model

// 代码画板包定义了 CodeCanvas 的核心数据结构
// ，CodeCanvas 是一款轻量级的代码性能分析和框架检测引擎
// 本版本仅专注于技术栈识别——不包含安全元数据。
// 常量定义
const (
	// 规则类型
	RuleTypeFramework = "framework"
	RuleTypeComponent = "component"

	// 代码所处的应用类别
	CategoryFrontend = "frontend"
	CategoryBackend  = "backend"
	CategoryDesktop  = "desktop"
	CategoryOther    = "other"
)

// AllCategory 代码类型的分类 前端 后端 桌面 其他
var AllCategory = []string{CategoryFrontend, CategoryBackend, CategoryDesktop, CategoryOther}

// DetectionInfo 框架与组件识别结果 包含已检测到的框架和组件的列表。
type DetectionInfo struct {
	Frameworks []DetectedItem `json:"frameworks"`
	Components []DetectedItem `json:"components"`
}

// DetectedItem  框架与组件识别结果代表了一项已检测到的技术项目（框架或组件）。
type DetectedItem struct {
	Name     string `json:"name"`     // 例如: "gin", "log4j-core", "wails"
	Type     string `json:"type"`     // "framework" 或 "component"
	Language string `json:"language"` // 例如: "Go", "Java", "JavaScript"
	Version  string `json:"version"`  // 版本字符串，可能为空
	Category string `json:"category"` // "frontend" | "backend" | "desktop"
	Evidence string `json:"evidence"` // 人类可读的检测原因
}

// LangInfo  某一编程语言或标记语言的详细统计数据。
type LangInfo struct {
	Name         string `json:"name"` // 例如: "Java", "YAML"
	Files        int    `json:"files"`
	CodeLines    int    `json:"codeLines"`
	CommentLines int    `json:"commentLines"`
	BlankLines   int    `json:"blankLines"`
}

// FrameworkMetadata 支持列表元数据 描述了 CodeCanvas 能够识别的一种框架。
type FrameworkMetadata struct {
	Name     string            `json:"name"`
	Language string            `json:"language"`
	Levels   map[string]string `json:"levels"` // 例如: {"L1": "pom.xml", "L2": "Application.java"}
}

// ComponentMetadata 描述了 CodeCanvas 能够识别的一种组件。
type ComponentMetadata struct {
	Name     string            `json:"name"`
	Language string            `json:"language"`
	Levels   map[string]string `json:"levels"` // 例如: {"L1": "pom.xml", "L2": "Application.java"}
}

// LangSummary 保存单一语言的统计结果
type LangSummary struct {
	Name    string
	Code    int64
	Comment int64
	Blank   int64
	Count   int64
}
