package canvas

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/winezer0/codecanvas/internal/analyzer"
	"github.com/winezer0/codecanvas/internal/engine"
	"github.com/winezer0/codecanvas/internal/model"
)

// Analyze performs a full analysis and returns a CanvasReport.
func Analyze(path string, rulesDir string) (*model.CanvasReport, error) {
	ctx := context.Background()

	// Analyze code profile
	az := analyzer.NewCodeAnalyzer()
	profile, index, err := az.AnalyzeCodeProfile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error analyzing code profile: %v", err)
	}

	// Create rule engine
	ruleEngine, err := engine.NewCanvasEngine(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("error loading rules: %v", err)
	}

	// 获取检测到的语言
	languages := make([]string, 0, len(profile.Languages))
	for _, lang := range profile.Languages {
		languages = append(languages, lang.Name)
	}

	// 扩展语言（例如 TSX -> JavaScript）以确保规则匹配
	languages = expandLanguages(languages)

	// 检测框架和组件
	detectionResult, err := ruleEngine.DetectFrameworks(ctx, index, languages)
	if err != nil {
		return nil, fmt.Errorf("error detecting frameworks and components: %v", err)
	}

	// 生成分析报告
	report := &model.CanvasReport{
		CodeProfile: *profile,
		Detection:   *detectionResult,
		Timestamp:   time.Now(),
	}
	return report, nil
}

// AnalyzeDirectory 对指定目录进行代码画板分析，返回分类后的结果。
// rulesDir 可选，指定自定义规则目录。如果为空，则仅使用内置规则。
func AnalyzeDirectory(path string, rulesDir string) (*model.AnalysisResult, error) {
	// 调用 canvas 包的核心分析函数
	// 注意：canvas.Analyze 返回的是 CanvasReport，包含了更详细的信息
	report, err := Analyze(path, rulesDir)
	if err != nil {
		return nil, err
	}

	// 转换语言列表为 Map 以便快速查找统计信息
	langStats := make(map[string]model.LanguageInfo)
	for _, l := range report.CodeProfile.Languages {
		langStats[l.Name] = l
	}

	result := &model.AnalysisResult{
		Languages:             report.CodeProfile.Languages,
		DesktopLanguages:      report.CodeProfile.DesktopLanguages,
		MainFrontendLanguages: getTopLanguages(report.CodeProfile.FrontendLanguages, langStats, nil, 3),
		FrontendLanguages:     report.CodeProfile.FrontendLanguages,
		MainBackendLanguages:  getTopLanguages(report.CodeProfile.BackendLanguages, langStats, nil, 3),
		BackendLanguages:      report.CodeProfile.BackendLanguages,
		Frameworks:            getUniqueItemNames(report.Detection.Frameworks),
		Components:            getUniqueItemNames(report.Detection.Components),
	}

	return result, nil
}

// getTopLanguages 根据代码行数和文件数对语言进行排序并返回前 N 个
func getTopLanguages(candidates []string, stats map[string]model.LanguageInfo, exclude []string, limit int) []string {
	// 过滤需要排除的语言
	excludeMap := make(map[string]bool)
	for _, e := range exclude {
		excludeMap[strings.ToLower(e)] = true
	}

	var validLangs []model.LanguageInfo
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

// getUniqueItemNames 提取去重后的 items (组件名或者框架名)名称 列表
func getUniqueItemNames(components []model.DetectedItem) []string {
	seen := make(map[string]bool)
	var languages []string

	for _, item := range components {
		name := item.Name
		if name == "" {
			continue // 跳过空语言（可选）
		}
		if !seen[name] {
			seen[name] = true
			languages = append(languages, name)
		}
	}

	return languages
}
