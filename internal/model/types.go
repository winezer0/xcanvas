package model

import "time"

// 代码画板包定义了 CodeCanvas 的核心数据结构
//，CodeCanvas 是一款轻量级的代码性能分析和框架检测引擎
// 本版本仅专注于技术栈识别——不包含安全元数据。

// 常量定义
const (
	// 检测结果的置信度水平
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"

	// 规则类型
	RuleTypeFramework = "framework"
	RuleTypeComponent = "component"

	// 代码所处的应用类别
	CategoryFrontend = "frontend"
	CategoryBackend  = "backend"
	CategoryDesktop  = "desktop"
)

// AnalysisResult 包含了 CodeCanvas 分析的完整结果
type AnalysisResult struct {
	// 语言信息列表
	LanguageInfos []LanguageInfo `json:"language_infos"`
	// 语言列表
	Languages []string `json:"languages"`
	// 桌面语言列表
	DesktopLanguages []string `json:"desktop_languages"`
	// 前端语言列表
	FrontendLanguages []string `json:"frontend_languages"`
	// 后端语言列表
	BackendLanguages []string `json:"backend_languages"`
	// 主要后端语言列表 (Top 3)
	MainBackendLanguages []string `json:"main_backend_languages"`
	// 主要前端语言列表 (Top 3)
	MainFrontendLanguages []string `json:"main_frontend_languages"`
	// 框架信息列表
	Frameworks []string `json:"frameworks"`
	// 组件信息列表
	Components []string `json:"components"`
}

type CodeProfile struct {
	Path              string         `json:"path"`
	TotalFiles        int            `json:"total_files"`
	TotalLines        int            `json:"total_lines"`
	ErrorFiles        int            `json:"error_files"`        // Number of files that failed to process
	LanguageInfos     []LanguageInfo `json:"language_infos"`     // 所有检测到的语言的完整列表
	FrontendLanguages []string       `json:"frontend_languages"` // 例如: ["TypeScript", "JavaScript"]
	BackendLanguages  []string       `json:"backend_languages"`  // 例如: ["Java", "Go"]
	DesktopLanguages  []string       `json:"desktop_languages"`  // 例如: ["C#", "C++"]
	Languages         []string       `json:"languages"`
	ExpandLanguages   []string       `json:"expand_languages"`
}

// LanguageInfo  某一编程语言或标记语言的详细统计数据。
type LanguageInfo struct {
	Name         string `json:"name"` // 例如: "Java", "YAML"
	Files        int    `json:"files"`
	CodeLines    int    `json:"code_lines"`
	CommentLines int    `json:"comment_lines"`
	BlankLines   int    `json:"blank_lines"`
}

// DetectionResult 框架与组件识别结果 包含已检测到的框架和组件的列表。
type DetectionResult struct {
	Frameworks []DetectedItem `json:"frameworks"`
	Components []DetectedItem `json:"components"`
}

// DetectedItem  框架与组件识别结果代表了一项已检测到的技术项目（框架或组件）。
type DetectedItem struct {
	Name       string `json:"name"`       // 例如: "gin", "log4j-core", "wails"
	Type       string `json:"type"`       // "framework" 或 "component"
	Language   string `json:"language"`   // 例如: "Go", "Java", "JavaScript"
	Version    string `json:"version"`    // 版本字符串，可能为空
	Category   string `json:"category"`   // "frontend" | "backend" | "desktop"
	Confidence string `json:"confidence"` // "high" | "medium" | "low"
	Evidence   string `json:"evidence"`   // 人类可读的检测原因
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

// FrameworkRuleDefinition 内部规则模型（对应 YAML 规则文件） 定义了如何检测框架或组件。 在启动时从 YAML 规则文件中加载。
type FrameworkRuleDefinition struct {
	Name     string              `yaml:"name"`
	Type     string              `yaml:"type"` // "framework" or "component"
	Language string              `yaml:"language"`
	Category string              `yaml:"category"` // 针对框架: "frontend"/"backend"; 针对组件: "frontend"/"backend"
	Levels   map[string]*RuleSet `yaml:"levels"`   // 键: "L1", "L2", "L3"
}

// RuleSet 针对一个检测级别，提供了组文件路径、内容关键词以及版本提取逻辑。
type RuleSet struct {
	Paths []string `yaml:"paths"`
	// 可选: "contains" 中的所有字符串必须出现在文件内容中（逻辑与）
	Contains []string `yaml:"contains,omitempty"`
	// 可选: 使用正则表达式从文件内容中提取版本
	ExtractVersionFromText *VersionExtractor `yaml:"extract_version_from_text,omitempty"`
	// 可选: 使用正则表达式从匹配的文件路径中提取版本
	ExtractVersionFromPath *VersionExtractor `yaml:"extract_version_from_path,omitempty"`
}

// VersionExtractor 定义了一个正则表达式模式，用于提取版本字符串 此模式必须包含一个捕获组；第一个组将用作版本号。
type VersionExtractor struct {
	Pattern string `yaml:"pattern"`
}

// CanvasReport 最终分析报告
type CanvasReport struct {
	CodeProfile CodeProfile     `json:"code_profile"`
	Detection   DetectionResult `json:"detection"`
	Timestamp   time.Time       `json:"timestamp"`
	Version     string          `json:"version"`
}

// FrontendLanguageSet 包含已知的前端相关语言。  用于填充 CodeProfile.FrontendLanguages 字段。
var FrontendLanguageSet = map[string]bool{
	"JavaScript": true,
	"TypeScript": true,
	"HTML":       true,
	"CSS":        true,
	"Vue":        true,
	"Svelte":     true,
	"JSX":        true,
	"TSX":        true,
	"SCSS":       true,
	"Less":       true,
	"Stylus":     true,
	"Handlebars": true,
	"Pug":        true,
}

// BackendLanguageSet 包含已知的后端/服务器端编程语言。 用于填充 CodeProfile.BackendLanguages 字段。
var BackendLanguageSet = map[string]bool{
	"Java":        true,
	"Python":      true,
	"Go":          true,
	"C#":          true,
	"PHP":         true,
	"Ruby":        true,
	"Kotlin":      true,
	"Scala":       true,
	"Rust":        true,
	"Swift":       true,
	"Objective-C": true,
	"Perl":        true,
	"Lua":         true,
	"Elixir":      true,
	"Groovy":      true,
	"Shell":       true, // often used in scripts/backends
	"PowerShell":  true,
}
