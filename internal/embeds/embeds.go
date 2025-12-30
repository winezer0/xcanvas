package embeds

import (
	"errors"
	"io"
	"io/fs"
	"strings"

	"github.com/winezer0/xcanvas/internal/frameembeds"
	"github.com/winezer0/xcanvas/internal/langembeds"
	"github.com/winezer0/xcanvas/internal/model"

	"gopkg.in/yaml.v3"
)

// EmbeddedFrameRules returns the default set of framework and component detection rules.
// Rules are embedded in the binary using the embed package and loaded from YAML files.
func EmbeddedFrameRules() []*model.Framework {
	var allRules []*model.Framework

	// Read all files from the embedded filesystem
	files, err := fs.Glob(frameembeds.FrameEmbedFS, "*.yml")
	if err != nil {
		// Should not happen in a valid build
		return []*model.Framework{}
	}

	for _, filename := range files {
		fileContent, err := frameembeds.FrameEmbedFS.ReadFile(filename)
		if err != nil {
			continue
		}

		// Try to parse as single-document array format first
		var rulesArray []*model.Framework
		if err := yaml.Unmarshal(fileContent, &rulesArray); err == nil {
			// Check if we actually got something valid (array of structs)
			// yaml.Unmarshal might succeed with empty array or zero values
			if len(rulesArray) > 0 && rulesArray[0].Name != "" {
				allRules = append(allRules, rulesArray...)
				continue
			}
		}

		// Fall back to multi-document format
		yamlReader := strings.NewReader(string(fileContent))
		decoder := yaml.NewDecoder(yamlReader)

		for {
			var rule model.Framework
			if err := decoder.Decode(&rule); err != nil {
				if err == io.EOF {
					break
				}
				// Skip malformed documents but continue with other files
				break
			}
			// Only append valid rules
			if rule.Name != "" {
				allRules = append(allRules, &rule)
			}
		}
	}

	return allRules
}

// EmbeddedLangRules 从 embed.FS 中加载所有 .yml 文件并解析为语言分类规则
// 支持两种 YAML 格式：
//  1. 单个文件包含一个 LangRule 数组（推荐）
//  2. 多文档 YAML 流（每个文档是一个规则）
func EmbeddedLangRules() map[string]model.Language {
	rules := make(map[string]model.Language)

	files, err := fs.Glob(langembeds.LanguageEmbedFS, "*.yml")
	if err != nil {
		return rules // 返回空 map
	}

	for _, filename := range files {
		content, err := langembeds.LanguageEmbedFS.ReadFile(filename)
		if err != nil {
			continue // 跳过无法读取的文件
		}

		// 尝试作为单文档数组解析
		var rulesArray []model.Language
		if err := yaml.Unmarshal(content, &rulesArray); err == nil && len(rulesArray) > 0 && rulesArray[0].Name != "" {
			for _, rule := range rulesArray {
				if rule.Name != "" {
					rules[strings.ToLower(rule.Name)] = rule
				}
			}
			continue
		}

		// 回退到多文档流解析
		decoder := yaml.NewDecoder(strings.NewReader(string(content)))
		for {
			var rule model.Language
			if err := decoder.Decode(&rule); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				// 遇到非 EOF 错误，放弃当前文件
				break
			}
			if rule.Name != "" {
				rules[strings.ToLower(rule.Name)] = rule
			}
		}
	}

	return rules
}
