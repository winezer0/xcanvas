package model

// AnalysisResult 包含了 CodeCanvas 分析的完整结果
type AnalysisResult struct {
	// 语言列表
	Languages []LanguageInfo `json:"languages"`
	// 桌面语言列表
	DesktopLanguages []string `json:"desktop_languages"`
	// 主要前端语言列表 (Top 3)
	MainFrontendLanguages []string `json:"main_frontend_languages"`
	// 前端语言列表
	FrontendLanguages []string `json:"frontend_languages"`
	// 主要后端语言列表 (Top 3)
	MainBackendLanguages []string `json:"main_backend_languages"`
	// 后端语言列表
	BackendLanguages []string `json:"backend_languages"`
	// 框架信息列表
	Frameworks []string `json:"frameworks"`
	// 组件信息列表
	Components []string `json:"components"`
}
