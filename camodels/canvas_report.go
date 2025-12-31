package camodels

import (
	"fmt"
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

// CanvasSimple 包含了 CodeCanvas 分析的完整结果
type CanvasSimple struct {
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

// ToSimpleReport 转换为简单的报告模式
func (report *CanvasReport) ToSimpleReport() *CanvasSimple {
	// 转换语言列表为 Map 以便快速查找统计信息
	langStats := languageInfosToMap(report.CodeProfile.LanguageInfos)
	result := &CanvasSimple{
		LanguageInfos:         report.CodeProfile.LanguageInfos,
		Languages:             report.CodeProfile.Languages,
		DesktopLanguages:      report.CodeProfile.DesktopLanguages,
		FrontendLanguages:     report.CodeProfile.FrontendLanguages,
		BackendLanguages:      report.CodeProfile.BackendLanguages,
		OtherLanguages:        report.CodeProfile.OtherLanguages,
		Frameworks:            getItemsWithVersions(report.Detection.Frameworks),
		Components:            getItemsWithVersions(report.Detection.Components),
		MainFrontendLanguages: getTopLanguages(report.CodeProfile.FrontendLanguages, langStats, nil, 3),
		MainBackendLanguages:  getTopLanguages(report.CodeProfile.BackendLanguages, langStats, nil, 3),
	}
	return result
}

// PrintCanvasReport outputs the analysis report in text format.
func (report *CanvasReport) PrintCanvasReport() {
	fmt.Println("CodeCanvas Analysis Report")
	fmt.Println("=========================")
	fmt.Printf("Path: %s\n", report.CodeProfile.Path)
	fmt.Printf("Total Files: %d\n", report.CodeProfile.TotalFiles)
	fmt.Printf("Total Lines: %d\n", report.CodeProfile.TotalLines)
	fmt.Println()

	// Frontend languages
	if len(report.CodeProfile.FrontendLanguages) > 0 {
		fmt.Println("Frontend LanguageInfos:")
		for _, lang := range report.CodeProfile.FrontendLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// Backend languages
	if len(report.CodeProfile.BackendLanguages) > 0 {
		fmt.Println("Backend LanguageInfos:")
		for _, lang := range report.CodeProfile.BackendLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// Desktop languages
	if len(report.CodeProfile.DesktopLanguages) > 0 {
		fmt.Println("Desktop LanguageInfos:")
		for _, lang := range report.CodeProfile.DesktopLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// Other languages
	if len(report.CodeProfile.OtherLanguages) > 0 {
		fmt.Println("Other LanguageInfos:")
		for _, lang := range report.CodeProfile.OtherLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// All languages (verbose only)
	fmt.Println("All LanguageInfos:")
	for _, lang := range report.CodeProfile.LanguageInfos {
		fmt.Printf("- %s: %d files, %d lines\n", lang.Name, lang.Files, lang.CodeLines)
	}
	fmt.Println()

	// Frameworks
	if len(report.Detection.Frameworks) > 0 {
		printDetectedItems("Detected Frameworks", report.Detection.Frameworks)
	} else {
		fmt.Printf("Detected Frameworks Is Empty !!!\n")
	}

	// Components
	if len(report.Detection.Components) > 0 {
		printDetectedItems("Detected Components", report.Detection.Components)
	} else {
		fmt.Printf("Detected Components Is Empty !!!\n")
	}

	fmt.Printf("Generated: %s\n", report.Timestamp.Format(time.RFC1123))
}
