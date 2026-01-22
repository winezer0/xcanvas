package langengine

import (
	"strings"

	"github.com/winezer0/xcanvas/internal/embeds"
	"github.com/winezer0/xcanvas/internal/model"
	"github.com/winezer0/xutils/logging"
)

// LangClassify 语言分类器的主结构体
// - langMap: 存储所有统一语言模型的映射，键为小写的语言名称

type LangClassify struct {
	langMap map[string]model.Language
}

// LanguageRules 加载embeds的默认规则
var LanguageRules = embeds.EmbeddedLangRules()

// NewLangClassifier 创建一个新的语言分类器实例
// 初始化分类器并加载所有语言规则
func NewLangClassifier() *LangClassify {
	c := &LangClassify{
		langMap: LanguageRules,
	}
	return c
}

// DetectCategories 检测给定语言的分类（前端/后端/桌面）
// 参数:
// - root: 项目根目录路径
// - langs: 语言信息列表
// 返回值:
// - frontend: 前端语言列表
// - backend: 后端语言列表
// - desktop: 桌面语言列表
// - other: 其他语言列表
// - all: 所有语言列表（去重）
func (c *LangClassify) DetectCategories(root string, langs []model.LangInfo) (frontend, backend, desktop, other, all, expand []string) {
	frontedSet := make(map[string]bool)
	backendSet := make(map[string]bool)
	desktopSet := make(map[string]bool)
	otherSet := make(map[string]bool)
	allSet := make(map[string]bool) // 用于去重所有语言

	deps := readPackageJSONDeps(root)
	for _, langInfo := range langs {
		name := strings.ToLower(langInfo.Name)
		allSet[langInfo.Name] = true
		// 检查是否有统一语言模型
		if langRule, ok := c.langMap[name]; !ok {
			// 没有规则，分类为other
			logging.Errorf("lang model not found for %s", name)
			otherSet[langInfo.Name] = true
			continue
		} else {
			// 应用动态分类规则
			cats := ApplyDynamicHeuristics(root, langRule, deps)
			for _, cat := range cats {
				// 根据分类结果添加到相应的集合
				switch cat {
				case model.CategoryFrontend:
					frontedSet[langInfo.Name] = true
				case model.CategoryBackend:
					backendSet[langInfo.Name] = true
				case model.CategoryDesktop:
					desktopSet[langInfo.Name] = true
				default:
					otherSet[langInfo.Name] = true
				}
			}
		}
	}

	// 提取结果（保持顺序无关，若需排序可加 sort.Strings）
	frontend = mapkeys(frontedSet)
	backend = mapkeys(backendSet)
	desktop = mapkeys(desktopSet)
	other = mapkeys(otherSet)
	all = mapkeys(allSet)
	expand = ExpandLanguages(all)
	return
}

// mapkeys 辅助函数：从 map[string]bool 提取所有 key
// 参数:
// - m: 输入的映射
// 返回值:
// - []string: 映射中所有的键组成的切片
func mapkeys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
