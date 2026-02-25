// Package engine 提供了 CodeCanvas 的框架和组件检测功能。
package frameengine

import (
	"fmt"
	"strings"

	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xutils/logging"
)

// formatVersion 格式化版本号，去除常见前缀和多余字符
func formatVersion(version string) string {
	if version == "" {
		return ""
	}

	// 去除常见的版本前缀
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, "=")
	version = strings.TrimSpace(version)

	// 去除构建元数据
	if idx := strings.Index(version, "+"); idx > 0 {
		version = version[:idx]
	}

	return version
}

// CanvasEngine 实现框架和组件检测功能。
type CanvasEngine struct {
	rules          []*camodels.Framework
	frameworkRules map[string]*camodels.Framework
	componentRules map[string]*camodels.Framework
}

// NewCanvasEngine 创建一个新的规则引擎实例，默认加载嵌入式规则。
// 如果提供了规则目录（rulesDir），则会从该目录加载规则，并将其与嵌入式规则合并，
// 其中用户定义的规则将覆盖具有相同名称的嵌入式规则。
func NewCanvasEngine(rulesDir string) (*CanvasEngine, error) {
	engine := &CanvasEngine{
		rules:          []*camodels.Framework{},
		frameworkRules: make(map[string]*camodels.Framework),
		componentRules: make(map[string]*camodels.Framework),
	}

	// 首先加载嵌入式规则
	engine.loadEmbeddedRules()

	// 如果提供了规则目录，则加载用户定义的规则并与嵌入式规则合并
	if rulesDir != "" {
		err := engine.loadRulesFromDirectory(rulesDir)
		if err != nil {
			logging.Errorf("load rules from rule dir (%s) occur error: %v", rulesDir, err)
			return engine, err
		}
	}

	return engine, nil
}

// DetectFrameworks 根据加载的规则检测给定目录中的框架和组件。
// 使用文件索引进行加速。
func (e *CanvasEngine) DetectFrameworks(index *camodels.FileIndex, languages []string) (*camodels.DetectionInfo, error) {
	result := &camodels.DetectionInfo{
		Frameworks: []camodels.DetectedItem{},
		Components: []camodels.DetectedItem{},
	}

	// 创建索引匹配器
	matcher := NewIndexMatcher(index)

	// 按检测到的语言过滤规则
	filteredRules := e.filterRulesByLanguages(languages)

	// 文件内容缓存
	fileContentCache := make(map[string][]byte)

	// 遍历所有规则，对每个框架进行检测
	for _, framework := range filteredRules {
		// 遍历框架的所有规则（OR关系）
		if matchFrame(matcher, framework.Rules, fileContentCache) {
			// 提取版本信息
			version := extractorVersion(matcher, framework.Versions, fileContentCache)
			// 规则匹配成功，创建检测结果
			item := camodels.DetectedItem{
				Name:     framework.Name,
				Type:     framework.Type,
				Language: framework.Language,
				Version:  version,
				Category: framework.Category,
				Evidence: fmt.Sprintf("FrameRule matched for %s", framework.Name),
			}
			// 根据规则类型添加到结果
			switch framework.Type {
			case camodels.RuleTypeFramework:
				result.Frameworks = append(result.Frameworks, item)
			case camodels.RuleTypeComponent:
				result.Components = append(result.Components, item)
			}
		}
	}

	return result, nil
}

// filterRulesByLanguages 过滤规则，只包含与检测到的语言匹配的规则。
func (e *CanvasEngine) filterRulesByLanguages(languages []string) []*camodels.Framework {
	var filtered []*camodels.Framework

	for _, rule := range e.rules {
		// 检查规则的语言是否在检测到的语言中
		for _, lang := range languages {
			if rule.Language == lang {
				filtered = append(filtered, rule)
				break
			}
		}
	}

	return filtered
}
