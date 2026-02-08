// Package engine 提供了 CodeCanvas 的框架和组件检测功能。
package frameengine

import (
	"fmt"
	"github.com/winezer0/xutils/logging"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/winezer0/xcanvas/internal/model"
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
	rules          []*model.Framework
	frameworkRules map[string]*model.Framework
	componentRules map[string]*model.Framework
}

// NewCanvasEngine 创建一个新的规则引擎实例，默认加载嵌入式规则。
// 如果提供了规则目录（rulesDir），则会从该目录加载规则，并将其与嵌入式规则合并，
// 其中用户定义的规则将覆盖具有相同名称的嵌入式规则。
func NewCanvasEngine(rulesDir string) (*CanvasEngine, error) {
	engine := &CanvasEngine{
		rules:          []*model.Framework{},
		frameworkRules: make(map[string]*model.Framework),
		componentRules: make(map[string]*model.Framework),
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
func (e *CanvasEngine) DetectFrameworks(index *model.FileIndex, languages []string) (*model.DetectionInfo, error) {
	result := &model.DetectionInfo{
		Frameworks: []model.DetectedItem{},
		Components: []model.DetectedItem{},
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
			item := model.DetectedItem{
				Name:     framework.Name,
				Type:     framework.Type,
				Language: framework.Language,
				Version:  version,
				Category: framework.Category,
				Evidence: fmt.Sprintf("FrameRule matched for %s", framework.Name),
			}
			// 根据规则类型添加到结果
			switch framework.Type {
			case model.RuleTypeFramework:
				result.Frameworks = append(result.Frameworks, item)
			case model.RuleTypeComponent:
				result.Components = append(result.Components, item)
			}
		}
	}

	return result, nil
}

// GetFileContentWithCache 读取文件内容，带缓存和大文件截断（最大 5MB，只读前 1MB）
// cache 是外部传入的 map[string][]byte，用于跨调用共享缓存
func GetFileContentWithCache(path string, cache map[string][]byte) ([]byte, error) {
	if content, ok := cache[path]; ok {
		return content, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err == nil && stat.Size() > 5*1024*1024 { // >5MB
		content, err := io.ReadAll(io.LimitReader(f, 1*1024*1024)) // 读前1MB
		if err != nil {
			return nil, err
		}
		cache[path] = content
		return content, nil
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	cache[path] = content
	return content, nil
}

// filterRulesByLanguages 过滤规则，只包含与检测到的语言匹配的规则。
func (e *CanvasEngine) filterRulesByLanguages(languages []string) []*model.Framework {
	var filtered []*model.Framework

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

// GetSupportedFrameworks 从规则中提取有关所有可检测框架的元数据。
func (e *CanvasEngine) GetSupportedFrameworks() []model.FrameworkMetadata {
	var frameworks []model.FrameworkMetadata

	for _, framework := range e.frameworkRules {
		// 提取规则信息
		levels := make(map[string]string)
		for i, rule := range framework.Rules {
			if len(rule.Paths) > 0 {
				levels[fmt.Sprintf("FrameRule%d", i+1)] = rule.Paths[0] // 使用第一个路径作为代表
			}
		}

		frameworkMeta := model.FrameworkMetadata{
			Name:     framework.Name,
			Language: framework.Language,
			Levels:   levels,
		}

		frameworks = append(frameworks, frameworkMeta)
	}

	// 按名称对框架进行排序
	sort.Slice(frameworks, func(i, j int) bool {
		return frameworks[i].Name < frameworks[j].Name
	})

	return frameworks
}

// GetSupportedComponents 从规则中提取有关所有可检测 组件 的元数据
func (e *CanvasEngine) GetSupportedComponents() []model.ComponentMetadata {
	var components []model.ComponentMetadata
	for _, component := range e.componentRules {
		// 提取规则信息
		levels := make(map[string]string)
		for i, rule := range component.Rules {
			if len(rule.Paths) > 0 {
				levels[fmt.Sprintf("FrameRule%d", i+1)] = rule.Paths[0] // 使用第一个路径作为代表
			}
		}
		componentMeta := model.ComponentMetadata{
			Name:     component.Name,
			Language: component.Language,
			Levels:   levels,
		}

		components = append(components, componentMeta)
	}
	// Sort components by name
	sort.Slice(components, func(i, j int) bool {
		return components[i].Name < components[j].Name
	})
	return components
}
