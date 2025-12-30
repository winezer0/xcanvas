package frameengine

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/winezer0/xcanvas/internal/embeds"
	"github.com/winezer0/xcanvas/internal/model"
	"gopkg.in/yaml.v3"
)

// loadEmbeddedRules 将默认的嵌入式规则加载到规则引擎中。
func (e *CanvasEngine) loadEmbeddedRules() {
	embeddedRules := embeds.EmbeddedFrameRules()

	// 将嵌入式规则添加到引擎中
	for _, rule := range embeddedRules {
		e.addRule(rule)
	}
}

// addRule 向引擎添加单个规则，替换具有相同名称的任何现有规则。
func (e *CanvasEngine) addRule(rule *model.Framework) {
	// 检查规则是否已存在
	existingIndex := -1
	for i, r := range e.rules {
		if r.Name == rule.Name && r.Type == rule.Type && r.Language == rule.Language {
			existingIndex = i
			break
		}
	}

	// 如果规则存在，则替换它
	if existingIndex != -1 {
		e.rules[existingIndex] = rule
	} else {
		// 否则，添加新规则
		e.rules = append(e.rules, rule)
	}

	// 更新规则映射
	switch rule.Type {
	case model.RuleTypeFramework:
		e.frameworkRules[rule.Name] = rule
	case model.RuleTypeComponent:
		e.componentRules[rule.Name] = rule
	}
}

// loadRulesFromDirectory 从给定目录加载所有 YAML 规则文件。
func (e *CanvasEngine) loadRulesFromDirectory(rulesDir string) error {
	// 读取目录中的所有 YAML 文件
	yamlFiles, err := filepath.Glob(filepath.Join(rulesDir, "*.yml"))
	if err != nil {
		return err
	}

	for _, yamlFile := range yamlFiles {
		rules, err := e.loadRulesFromFile(yamlFile)
		if err != nil {
			return err
		}

		// 使用 addRule 方法将规则添加到引擎中以处理规则合并
		for _, rule := range rules {
			e.addRule(rule)
		}
	}

	return nil
}

// loadRulesFromFile 从单个 YAML 文件加载规则，支持单文档数组格式和多文档格式。
func (e *CanvasEngine) loadRulesFromFile(filePath string) ([]*model.Framework, error) {
	// 读取 YAML 文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 首先尝试解析为单文档数组格式
	var rulesArray []*model.Framework
	if err := yaml.Unmarshal(data, &rulesArray); err == nil {
		// 成功解析为数组格式
		return rulesArray, nil
	}

	// 如果数组格式解析失败，尝试多文档格式
	yamlReader := strings.NewReader(string(data))
	decoder := yaml.NewDecoder(yamlReader)

	var rules []*model.Framework

	// 解码 YAML 文件中的每个文档
	for {
		var rule model.Framework
		err := decoder.Decode(&rule)
		if err != nil {
			if err == io.EOF {
				// 已到达文件末尾
				break
			}
			return nil, err
		}

		// 将解码后的规则添加到列表中
		rules = append(rules, &rule)
	}

	return rules, nil
}
