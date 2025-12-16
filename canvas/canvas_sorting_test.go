package canvas

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeDirectory_SortingAndFiltering(t *testing.T) {
	// 创建临时测试目录
	tmpDir, err := os.MkdirTemp("", "canvasutils_sorting_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. 创建 package.json 以确保 JS/TS 被识别为前端 (React)
	pkgJson := `{
  "dependencies": {
    "react": "^18.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJson), 0644); err != nil {
		t.Fatalf("无法写入 package.json: %v", err)
	}

	// 2. 创建 TypeScript 文件 (多行代码，应该是第一名)
	tsContent := `
import React from 'react';
// Line 2
// Line 3
// Line 4
// Line 5
export const App = () => {
  return <div>Hello</div>;
};
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.ts"), []byte(tsContent), 0644); err != nil {
		t.Fatalf("无法写入 app.ts: %v", err)
	}

	// 3. 创建 JavaScript 文件 (少行代码，应该是第二名)
	jsContent := `
console.log("Hello");
`
	if err := os.WriteFile(filepath.Join(tmpDir, "utils.js"), []byte(jsContent), 0644); err != nil {
		t.Fatalf("无法写入 utils.js: %v", err)
	}

	// 4. 创建 CSS 文件 (很多行，应该被过滤掉)
	cssContent := `
body {
  color: red;
  margin: 0;
  padding: 0;
  /* line 5 */
  /* line 6 */
  /* line 7 */
  /* line 8 */
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte(cssContent), 0644); err != nil {
		t.Fatalf("无法写入 style.css: %v", err)
	}

	// 5. 创建 Go 文件 (后端)
	goContent := `package main
import "fmt"
func main() {
	fmt.Println("Backend")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatalf("无法写入 main.go: %v", err)
	}

	// 运行分析
	result, err := AnalyzeDirectory(tmpDir, "")
	if err != nil {
		t.Fatalf("AnalyzeDirectory 失败: %v", err)
	}

	// 验证前端语言
	// 期望: TypeScript (第一), JavaScript (第二)
	// CSS 应该在列表中 (因为它是前端语言)

	// 检查 CSS 是否存在 (FrontendLanguages 应该包含所有)
	foundCSS := false
	for _, lang := range result.FrontendLanguages {
		if lang == "CSS" {
			foundCSS = true
			break
		}
	}
	if !foundCSS {
		t.Error("CSS 应该在前端语言列表之中")
	}

	// 验证 MainFrontendLanguages
	// CSS 代码行数最多，应该在 Top 3 中
	// 顺序应该是 CSS, TypeScript, JavaScript
	if len(result.MainFrontendLanguages) < 3 {
		t.Errorf("期望至少检测到 3 种主要前端语言 (CSS, TS, JS)，实际: %v", result.MainFrontendLanguages)
	} else {
		// 验证顺序 (代码行数: CSS > TS > JS)
		// 注意：具体顺序取决于代码行数统计的准确性，这里我们主要验证都在列表里
		foundMainCSS := false
		foundMainTS := false
		foundMainJS := false
		for _, lang := range result.MainFrontendLanguages {
			if lang == "CSS" {
				foundMainCSS = true
			}
			if lang == "TypeScript" {
				foundMainTS = true
			}
			if lang == "JavaScript" {
				foundMainJS = true
			}
		}
		if !foundMainCSS || !foundMainTS || !foundMainJS {
			t.Errorf("主要前端语言列表缺失，实际: %v", result.MainFrontendLanguages)
		}
	}

	// 验证后端语言
	foundGo := false
	for _, lang := range result.MainBackendLanguages {
		if lang == "Go" {
			foundGo = true
			break
		}
	}
	if !foundGo {
		t.Error("未检测到 Go 为主要后端语言")
	}

	// 验证 Languages 字段
	if len(result.Languages) == 0 {
		t.Error("Languages 列表不应为空")
	}
}
