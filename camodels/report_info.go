package camodels

import (
	"sort"
	"strings"
	"time"

	"github.com/winezer0/xcanvas/internal/model"
)

// CodeProfile 代码轮廓
type CodeProfile struct {
	Path              string           `json:"path"`
	TotalFiles        int              `json:"totalFiles"`
	TotalLines        int              `json:"totalLines"`
	ErrorFiles        int              `json:"errorFiles"`        // Number of files that failed to process
	LanguageInfos     []model.LangInfo `json:"languageInfos"`     // 所有检测到的语言的完整列表
	FrontendLanguages []string         `json:"frontendLanguages"` // 例如: ["TypeScript", "JavaScript"]
	BackendLanguages  []string         `json:"backendLanguages"`  // 例如: ["Java", "Go"]
	DesktopLanguages  []string         `json:"desktopLanguages"`  // 例如: ["C#", "C++"]
	OtherLanguages    []string         `json:"otherLanguages"`    // 例如: ["JSON", "YAML"]
	Languages         []string         `json:"languages"`
	Expands           []string         `json:"expands"`
}

// CanvasReport 最终分析报告
type CanvasReport struct {
	CodeProfile CodeProfile         `json:"codeProfile"`
	Detection   model.DetectionInfo `json:"detection"`
	Timestamp   time.Time           `json:"timestamp"`
	Version     string              `json:"version"`
}

// CanvasSimple 包含了 CodeCanvas 分析的完整结果
type CanvasSimple struct {
	// 语言信息列表
	LanguageInfos []model.LangInfo `json:"languageInfos"`
	// 语言列表
	Languages []string `json:"languages"`
	// 前端语言列表
	FrontendLanguages []string `json:"frontendLanguages"`
	// 后端语言列表
	BackendLanguages []string `json:"backendLanguages"`
	// 桌面语言列表
	DesktopLanguages []string `json:"desktopLanguages"`
	// 其他语言列表
	OtherLanguages []string `json:"otherLanguages"`
	// 主要后端语言列表 (Top 3)
	MainBackendLanguages []string `json:"mainBackendLanguages"`
	// 主要前端语言列表 (Top 3)
	MainFrontendLanguages []string `json:"mainFrontendLanguages"`
	// 框架信息列表，名称到版本的映射
	Frameworks map[string]string `json:"frameworks"`
	// 组件信息列表，名称到版本的映射
	Components map[string]string `json:"components"`
	// 记录分析文件数量
	TotalFiles int `json:"totalFiles"`
}

// ToSimpleReport 转换为简单的报告模式
func (report *CanvasReport) ToSimpleReport() *CanvasSimple {
	// 转换语言列表为 Map 以便快速查找统计信息
	langStats := languageInfosToMap(report.CodeProfile.LanguageInfos)
	result := &CanvasSimple{
		TotalFiles:            report.CodeProfile.TotalFiles,
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

func languageInfosToMap(languageInfos []model.LangInfo) map[string]model.LangInfo {
	langStats := make(map[string]model.LangInfo)
	for _, l := range languageInfos {
		langStats[l.Name] = l
	}
	return langStats
}

// getItemsWithVersions 提取去重后的 items (组件名或者框架名)及其版本，返回名称到版本的映射
func getItemsWithVersions(items []model.DetectedItem) map[string]string {
	result := make(map[string]string)

	for _, item := range items {
		name := item.Name
		if name == "" {
			continue // 跳过空名称
		}
		// 如果版本为空，使用空字符串
		version := item.Version
		// 如果已经存在，不覆盖，保留第一个匹配的版本
		if _, exists := result[name]; !exists {
			result[name] = version
		}
	}

	return result
}

// getTopLanguages 根据代码行数和文件数对语言进行排序并返回前 N 个
func getTopLanguages(candidates []string, stats map[string]model.LangInfo, exclude []string, limit int) []string {
	// 过滤需要排除的语言
	excludeMap := make(map[string]bool)
	for _, e := range exclude {
		excludeMap[strings.ToLower(e)] = true
	}

	var validLangs []model.LangInfo
	for _, name := range candidates {
		if excludeMap[strings.ToLower(name)] {
			continue
		}
		if info, ok := stats[name]; ok {
			validLangs = append(validLangs, info)
		}
	}

	// 排序：优先代码行数，其次文件数
	sort.Slice(validLangs, func(i, j int) bool {
		if validLangs[i].CodeLines != validLangs[j].CodeLines {
			return validLangs[i].CodeLines > validLangs[j].CodeLines
		}
		return validLangs[i].Files > validLangs[j].Files
	})

	// 取前 N 个
	var result []string
	count := 0
	for _, l := range validLangs {
		if count >= limit {
			break
		}
		result = append(result, l.Name)
		count++
	}
	return result
}
