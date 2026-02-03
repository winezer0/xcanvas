package canvas

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeDirectory(t *testing.T) {
	// 创建临时测试目录
	tmpDir, err := os.MkdirTemp("", "canvasutils_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建一个模拟的 Go 后端项目
	// 1. go.mod (Gin 框架)
	goModContent := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("无法写入 go.mod: %v", err)
	}

	// 2. main.go (Go 语言文件)
	mainGoContent := `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.Run()
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("无法写入 main.go: %v", err)
	}

	// 运行分析
	result, err := Analyze(tmpDir, "")
	if err != nil {
		t.Fatalf("Analyze 失败: %v", err)
	}

	// 验证结果

	// 1. 验证后端语言
	foundGo := false
	for _, lang := range result.CodeProfile.BackendLanguages {
		if lang == "Go" {
			foundGo = true
			break
		}
	}
	if !foundGo {
		t.Error("未检测到 Go 为后端语言")
	}

	// 2. 验证框架 (Gin)
	foundGin := false
	for _, fw := range result.Detection.Frameworks {
		if fw.Name == "Gin" {
			foundGin = true
			break
		}
	}
	if !foundGin {
		t.Error("未检测到 Gin 框架")
	}

	// 3. 验证前端语言 (应为空)
	if len(result.CodeProfile.FrontendLanguages) > 0 {
		t.Errorf("检测到前端语言列表: %v", result.CodeProfile.FrontendLanguages)
	}
}
