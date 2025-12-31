package camodels

import (
	"time"

	"github.com/winezer0/xcanvas/internal/model"
)

// CodeProfile 代码轮廓
type CodeProfile struct {
	Path              string           `json:"path"`
	TotalFiles        int              `json:"total_files"`
	TotalLines        int              `json:"total_lines"`
	ErrorFiles        int              `json:"error_files"`        // Number of files that failed to process
	LanguageInfos     []model.LangInfo `json:"language_infos"`     // 所有检测到的语言的完整列表
	FrontendLanguages []string         `json:"frontend_languages"` // 例如: ["TypeScript", "JavaScript"]
	BackendLanguages  []string         `json:"backend_languages"`  // 例如: ["Java", "Go"]
	DesktopLanguages  []string         `json:"desktop_languages"`  // 例如: ["C#", "C++"]
	OtherLanguages    []string         `json:"other_languages"`    // 例如: ["JSON", "YAML"]
	Languages         []string         `json:"languages"`
	Expands           []string         `json:"expands"`
}

// CanvasReport 最终分析报告
type CanvasReport struct {
	CodeProfile CodeProfile         `json:"code_profile"`
	Detection   model.DetectionInfo `json:"detection"`
	Timestamp   time.Time           `json:"timestamp"`
	Version     string              `json:"version"`
}

// AnalysisResult 包含了 CodeCanvas 分析的完整结果
type AnalysisResult struct {
	// 语言信息列表
	LanguageInfos []model.LangInfo `json:"language_infos"`
	// 语言列表
	Languages []string `json:"languages"`
	// 前端语言列表
	FrontendLanguages []string `json:"frontend_languages"`
	// 后端语言列表
	BackendLanguages []string `json:"backend_languages"`
	// 桌面语言列表
	DesktopLanguages []string `json:"desktop_languages"`
	// 其他语言列表
	OtherLanguages []string `json:"other_languages"`
	// 主要后端语言列表 (Top 3)
	MainBackendLanguages []string `json:"main_backend_languages"`
	// 主要前端语言列表 (Top 3)
	MainFrontendLanguages []string `json:"main_frontend_languages"`
	// 框架信息列表，名称到版本的映射
	Frameworks map[string]string `json:"frameworks"`
	// 组件信息列表，名称到版本的映射
	Components map[string]string `json:"components"`
}
