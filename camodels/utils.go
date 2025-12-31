package camodels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/winezer0/xcanvas/internal/model"
)

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

func languageInfosToMap(languageInfos []model.LangInfo) map[string]model.LangInfo {
	langStats := make(map[string]model.LangInfo)
	for _, l := range languageInfos {
		langStats[l.Name] = l
	}
	return langStats
}

func printDetectedItems(title string, items []model.DetectedItem) {
	fmt.Println(title + ":")

	// Group by category
	byCategory := make(map[string][]model.DetectedItem)
	for _, item := range items {
		byCategory[item.Category] = append(byCategory[item.Category], item)
	}

	categories := model.AllCategory
	for _, cat := range categories {
		printCategoryItems(cat, byCategory[cat])
	}

	// Handle any other categories
	for cat, items := range byCategory {
		isKnown := false
		for _, known := range categories {
			if cat == known {
				isKnown = true
				break
			}
		}
		if !isKnown && len(items) > 0 {
			printCategoryItems(cat, items)
		}
	}
	fmt.Println()
}

func printCategoryItems(category string, items []model.DetectedItem) {
	if len(items) > 0 {
		fmt.Printf("  [%s]\n", category)
		for _, item := range items {
			fmt.Printf("  - %s (%s)\n", item.Name, item.Language)
			if item.Version != "" {
				fmt.Printf("    Version: %s\n", item.Version)
			}
			if item.Evidence != "" {
				fmt.Printf("    Evidence: %s\n", item.Evidence)
			}
		}
	}
}
