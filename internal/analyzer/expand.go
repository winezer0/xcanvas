package analyzer

// expandLanguages 在给定的语言列表中，自动补充关联语言，以确保语义完整性。
// 例如：
// - TypeScript/TSX/JSX/Vue -> JavaScript (确保能匹配 JS 生态的规则)
// - SCSS/Less -> CSS (确保能匹配 CSS 规则)
// - Kotlin -> Java (确保能匹配 Java/JVM 生态规则)
// - C++ -> C (C++ 项目通常也包含 C 代码或库)
func expandLanguages(langs []string) []string {
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
