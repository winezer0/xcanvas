package langengine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/winezer0/xcanvas/internal/model"
)

// ApplyDynamicHeuristics 应用动态分类规则对语言进行分类
// 参数:
// - root: 项目根目录路径
// - lang: 统一语言模型
// - deps: 项目依赖映射
// 返回值:
// - string: 分类结果（frontend/backend/desktop/other）
func ApplyDynamicHeuristics(root string, lang model.Language, deps map[string]bool) []string {
	baseRes := []string{lang.Category}
	if len(lang.Dynamic) == 0 {
		return baseRes
	}

	// 首先检查动态分类规则
	for _, dynamic := range lang.Dynamic {
		// 检查依赖条件
		if len(dynamic.Dependencies) > 0 {
			for _, dep := range dynamic.Dependencies {
				if deps[strings.ToLower(dep)] {
					baseRes = append(baseRes, dynamic.Category)
				}
			}
		}
		// 检查文件模式条件
		if len(dynamic.FilePatterns) > 0 {
			for _, pattern := range dynamic.FilePatterns {
				matches, _ := filepath.Glob(filepath.Join(root, pattern))
				if len(matches) > 0 {
					baseRes = append(baseRes, dynamic.Category)
				}
			}
		}
	}

	// 如果没有匹配到动态规则，则使用默认分类
	return baseRes
}

// readPackageJSONDeps 从package.json读取项目依赖，用于JavaScript/TypeScript分类
// 参数:
// - root: 项目根目录路径
// 返回值:
// - map[string]bool: 依赖包名称映射（小写）
func readPackageJSONDeps(root string) map[string]bool {
	res := map[string]bool{}
	p := filepath.Join(root, "package.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return res
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	// 检查不同类型的依赖
	for _, key := range []string{"dependencies", "devDependencies", "peerDependencies"} {
		v, ok := m[key]
		if ok {
			mm, _ := v.(map[string]any)
			for k := range mm {
				res[strings.ToLower(k)] = true
			}
		}
	}
	return res
}

// ExpandLanguages 在给定的语言列表中，自动补充关联语言，以确保语义完整性。
// 例如：
// - TypeScript/TSX/JSX/Vue -> JavaScript (确保能匹配 JS 生态的规则)
// - SCSS/Less -> CSS (确保能匹配 CSS 规则)
// - Kotlin -> Java (确保能匹配 Java/JVM 生态规则)
// - C++ -> C (C++ 项目通常也包含 C 代码或库)
func ExpandLanguages(langs []string) []string {
	seen := make(map[string]bool)
	for _, l := range langs {
		seen[l] = true
	}

	// 辅助函数：如果语言不存在则添加
	add := func(newLang string) {
		if !seen[newLang] {
			langs = append(langs, newLang)
			seen[newLang] = true
		}
	}

	// 1. JavaScript 生态系统
	// Vue, React (JSX/TSX), TypeScript 都属于 JS 生态
	if seen["TypeScript"] || seen["TSX"] || seen["JSX"] || seen["Vue"] {
		add("JavaScript")
	}

	// 2. CSS 生态系统
	// 预处理器文件通常也意味着 CSS 规则适用
	if seen["SCSS"] || seen["Less"] {
		add("CSS")
	}

	// 3. Java/JVM 生态系统
	// Kotlin 通常与 Java 库/框架（如 Spring）混用
	if seen["Kotlin"] {
		add("Java")
	}

	// 4. C/C++ 生态系统
	// C++ 往往包含或链接 C 代码
	if seen["C++"] {
		add("C")
	}

	return langs
}
