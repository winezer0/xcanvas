package frameengine

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/winezer0/xcanvas/internal/model"
)

// IndexMatcher 提供基于索引的文件查找功能
type IndexMatcher struct {
	Index *model.FileIndex
}

// NewIndexMatcher 创建一个新的索引匹配器
func NewIndexMatcher(index *model.FileIndex) *IndexMatcher {
	return &IndexMatcher{Index: index}
}

// FindFiles 使用索引查找匹配的文件。
// pattern 支持:
// 1. 精确相对路径 (e.g., "/package.json")
// 2. 文件名匹配 (e.g., "package.json", "*.json")
// 3. 递归通配符 (e.g., "**/*.go", "src/**/*.js")
func (m *IndexMatcher) FindFiles(pattern string) ([]string, error) {
	var results []string

	// Case 1: 精确相对路径 (e.g. "/package.json")
	if strings.HasPrefix(pattern, "/") {
		target := strings.TrimPrefix(pattern, "/")
		// 检查索引中是否存在 - 使用小写文件名作为键
		for _, idx := range m.Index.NameMap[strings.ToLower(path.Base(target))] {
			f := m.Index.Files[idx]
			// 不区分大小写的路径比较
			if strings.EqualFold(f, target) {
				results = append(results, filepath.Join(m.Index.RootDir, f))
			}
		}
		// 如果索引中找不到，但模式是绝对的，可能我们应该信任模式？
		// 但这是基于索引的，所以我们只返回索引中的。
		return results, nil
	}

	// Case 2: 文件名 (e.g. "package.json")
	// 如果不包含路径分隔符和通配符，则自动匹配任意目录下的该文件
	if !strings.Contains(pattern, "/") && !strings.Contains(pattern, "*") {
		// 自动添加 **/ 前缀，匹配任意目录下的该文件
		// 直接遍历所有文件，检查文件名是否匹配
		for _, fileRelPath := range m.Index.Files {
			if strings.EqualFold(path.Base(fileRelPath), pattern) {
				results = append(results, filepath.Join(m.Index.RootDir, fileRelPath))
			}
		}
		return results, nil
	}

	// Case 3: 后缀匹配 (e.g. "*.json") - 优化
	if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "/") {
		// 使用小写扩展名作为ExtensionMap的键，实现不区分大小写的后缀匹配
		ext := strings.ToLower(pattern[1:]) // ".json"
		if indices, ok := m.Index.ExtensionMap[ext]; ok {
			for _, idx := range indices {
				results = append(results, filepath.Join(m.Index.RootDir, m.Index.Files[idx]))
			}
		}
		return results, nil
	}

	// Case 5: 目录匹配 (模式以 / 结尾)
	if strings.HasSuffix(pattern, "/") {
		// 使用小写进行比较，实现不区分大小写的目录匹配
		patternLower := strings.ToLower(pattern)
		for _, fileRelPath := range m.Index.Files {
			if strings.HasPrefix(strings.ToLower(fileRelPath), patternLower) {
				results = append(results, filepath.Join(m.Index.RootDir, fileRelPath))
			}
		}
		return results, nil
	}

	// Case 4: 通用 glob 匹配 (e.g. "src/**/*.ts")
	// 这需要遍历索引中的所有文件路径进行匹配
	// 这是一个 O(N) 操作，其中 N 是文件总数，比磁盘 I/O 快得多。
	for _, fileRelPath := range m.Index.Files {
		// filepath.Match 不支持 **，我们需要支持它。
		// 这里我们假设 pattern 遵循 gitignore 风格或 standard glob。
		// 为了简单起见，我们使用 filepath.Match 对每部分进行匹配，或者使用正则。
		// Go 的 filepath.Match 不支持递归 **。
		// 为了支持 **，我们可以使用第三方库，或者简单的实现：
		// 如果 pattern 包含 **，我们将 ** 替换为 .* 并使用正则。

		matched, err := matchPath(pattern, fileRelPath)
		if err == nil && matched {
			results = append(results, filepath.Join(m.Index.RootDir, fileRelPath))
		}
	}

	return results, nil
}

// matchPath 简单的路径匹配，支持 **
func matchPath(pattern, name string) (bool, error) {
	// 确保模式使用正斜杠以匹配 FileIndex 约定
	pattern = filepath.ToSlash(pattern)

	// 优化：如果 pattern 是 "**/package.json" 且 name 是 "package.json"，直接匹配
	if pattern == "**/package.json" && name == "package.json" {
		return true, nil
	}
	if pattern == "**/*.go" && strings.HasSuffix(name, ".go") {
		return true, nil
	}
	if pattern == "**/*.js" && strings.HasSuffix(name, ".js") {
		return true, nil
	}
	if pattern == "**/*.ts" && strings.HasSuffix(name, ".ts") {
		return true, nil
	}
	if pattern == "**/*.jsx" && strings.HasSuffix(name, ".jsx") {
		return true, nil
	}
	if pattern == "**/*.tsx" && strings.HasSuffix(name, ".tsx") {
		return true, nil
	}

	// 简单的 ** 支持
	if strings.Contains(pattern, "**") {
		// 简化的 ** 处理
		// 1. 如果 pattern 以 **/ 开头，检查 name 是否匹配剩余部分
		if strings.HasPrefix(pattern, "**/") {
			remaining := pattern[3:]
			if remaining == name {
				return true, nil
			}
			if strings.Contains(name, "/") {
				// 检查 name 的最后部分是否匹配 remaining
				lastPart := name[strings.LastIndex(name, "/")+1:]
				return path.Match(strings.ToLower(remaining), strings.ToLower(lastPart))
			}
			return false, nil
		}

		// 2. 如果 pattern 包含 /**/，检查是否存在该目录结构
		if strings.Contains(pattern, "/**/") {
			parts := strings.Split(pattern, "/**/")
			if len(parts) != 2 {
				return false, nil
			}
			prefix := parts[0]
			suffix := parts[1]

			// 检查 name 是否以 prefix 开头和 suffix 结尾
			return strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix), nil
		}
	}

	// 对于非 ** 模式，将 pattern 和 name 都转换为小写后再匹配
	return path.Match(strings.ToLower(pattern), strings.ToLower(name))
}
