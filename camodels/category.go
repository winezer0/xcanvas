package camodels

// 代码画板包定义了 CodeCanvas 的核心数据结构
// CodeCanvas 是一款轻量级的代码性能分析和框架检测引擎
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
