package frameengine

import (
	"github.com/winezer0/xutils/utils"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/winezer0/xcanvas/internal/model"
	"github.com/winezer0/xutils/logging"
)

// containsAllKeywords 检查文件内容是否包含所有必需的关键字。
// 如果 ignoreCase 为 true，则进行大小写不敏感匹配；否则区分大小写。
func containsAllKeywords(content []byte, keys []string, ignoreCase bool) bool {
	if len(content) == 0 || len(keys) == 0 {
		return false
	}

	contentStr := string(content)
	if ignoreCase {
		contentStr = strings.ToLower(contentStr)
		for i, kw := range keys {
			keys[i] = strings.ToLower(kw)
		}
	}

	for _, kw := range keys {
		if !strings.Contains(contentStr, kw) {
			return false
		}
	}
	return true
}

// extractVersion 使用给定的正则表达式列表从文件内容中提取版本号
// 按顺序尝试每个正则表达式，第一个成功匹配且包含捕获组的结果将被用作版本号
func extractVersion(content []byte, patterns []string) string {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}
	return ""
}

// matchFrame 检查 rules 中是否有任意一条规则被满足。
// 规则满足条件 = 所有 Paths 存在 AND 所有 FileContents 条件满足。
// 返回 true 表示至少有一条规则匹配成功。
func matchFrame(matcher *IndexMatcher, rules []model.FrameRule, fileContentCache map[string][]byte) bool {
	for _, rule := range rules {
		if len(rule.Paths) == 0 && len(rule.FileContents) == 0 {
			logging.Errorf("match rules not has any match content: %s", utils.ToJSON(rule))
			continue
		}

		// 1. 检查 Paths（所有路径必须存在，AND）
		pathsMatch := true // 假设全部存在
		if len(rule.Paths) > 0 {
			// 如果 Paths 不为空 需要先匹配paths列表 判断需要的文件路径是否都存在
			for _, path := range rule.Paths {
				matches, _ := matcher.FindFiles(filepath.ToSlash(path))
				if len(matches) == 0 {
					pathsMatch = false
					break // 存在path缺失即失败
				}
			}

			// 如果 Paths 不满足，跳过此规则
			if !pathsMatch {
				continue
			}
		}

		// 2. 检查 FileContents（每个 pattern 必须有至少一个文件包含其所有关键字，AND across patterns）
		fileMatch := true // 假设全部满足
		if len(rule.FileContents) > 0 {
			for filePattern, fileKeys := range rule.FileContents {
				findFiles, _ := matcher.FindFiles(filePattern)
				if len(findFiles) == 0 {
					fileMatch = false
					break // 没有文件匹配此 pattern，失败
				}

				// 检查是否存在至少一个文件包含所有关键字
				oneFileMatches := false
				for _, path := range findFiles {
					content, err := GetFileContentWithCache(path, fileContentCache)
					if err != nil {
						continue
					}
					if containsAllKeywords(content, fileKeys, true) {
						oneFileMatches = true
						break
					}
				}

				if !oneFileMatches {
					fileMatch = false
					break // 此 pattern 无文件满足，失败
				}
				// 否则继续检查下一个 pattern
			}
		}

		// 如果当前规则完全匹配（Paths + FileContents），立即返回 true
		if pathsMatch && fileMatch {
			return true
		}
	}

	// 所有规则都不匹配
	return false
}

func extractorVersion(matcher *IndexMatcher, versionExtractors []model.VersionExtractor, fileContentCache map[string][]byte) string {
	version := ""
	// 使用框架/组件级版本提取规则
	for _, versionExtractor := range versionExtractors {
		// 找到所有匹配该模式的文件
		findFiles, _ := matcher.FindFiles(versionExtractor.FilePattern)
		if len(findFiles) == 0 {
			// 没有匹配的文件，跳过此提取规则
			continue
		}

		// 检查所有匹配的文件，直到找到版本号
		for _, path := range findFiles {
			content, err := GetFileContentWithCache(path, fileContentCache)
			if err != nil {
				// 无法读取文件，尝试下一个文件
				continue
			}

			// 按顺序尝试每个正则表达式，直到找到匹配的版本号
			for _, pattern := range versionExtractor.Patterns {
				// 使用正则表达式提取版本号
				re, err := regexp.Compile(pattern)
				if err != nil {
					// 正则表达式无效，尝试下一个
					continue
				}

				matches := re.FindSubmatch(content)
				if len(matches) > 1 {
					// 找到匹配 并格式化版本号
					version = formatVersion(string(matches[1]))
					if len(version) > 0 {
						// 停止遍历匹配方法
						break
					}
				}
			}

			// 如果从文件内容未找到版本号，尝试从文件名提取
			for _, pattern := range versionExtractor.Patterns {
				// 使用正则表达式从文件名提取版本号
				re, err := regexp.Compile(pattern)
				if err != nil {
					// 正则表达式无效，尝试下一个
					continue
				}

				matches := re.FindStringSubmatch(path)
				if len(matches) > 1 {
					// 找到匹配 并 格式化版本号，去除 ^、~、= 等前缀和空格
					version = formatVersion(matches[1])
					if len(version) > 0 {
						// 停止遍历匹配方法
						break
					}
				}
			}

			// 如果找到版本号，停止遍历
			if len(version) > 0 {
				break
			}
		}

		// 如果找到版本号，停止遍历
		if len(version) > 0 {
			break
		}
	}
	return version
}
