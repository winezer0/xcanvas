package frameengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/winezer0/xcanvas/internal/model"
)

// TestEmbeddedRules verifies that every embedded rule can be triggered by a minimal test case.
func TestEmbeddedRules(t *testing.T) {
	// Initialize engine with embedded rules
	e, err := NewCanvasEngine("")
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}

	if len(e.rules) == 0 {
		t.Fatal("No rules loaded from embedded assets")
	}

	for _, framework := range e.rules {
		t.Run(fmt.Sprintf("%s-%s", framework.Language, framework.Name), func(t *testing.T) {
			// Create a temp directory for this rule
			tmpDir, err := os.MkdirTemp("", "rule_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Setup test environment based on the rule
			if !setupTestEnv(t, tmpDir, framework) {
				t.Skipf("Skipping rule %s: could not setup test environment (complex rule?)", framework.Name)
			}

			// Run detection
			ctx := context.Background()
			// We must provide the language the rule expects, otherwise it might skip
			languages := []string{framework.Language}

			// Build file index
			index, err := buildTestIndex(tmpDir)
			if err != nil {
				t.Fatalf("Failed to build file index: %v", err)
			}

			result, err := e.DetectFrameworks(ctx, index, languages)
			if err != nil {
				t.Fatalf("DetectFrameworks failed: %v", err)
			}

			// Verify detection
			found := false
			var items []model.DetectedItem
			if framework.Type == "framework" {
				items = result.Frameworks
			} else {
				items = result.Components
			}

			for _, item := range items {
				if item.Name == framework.Name {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Framework %s was not detected. Created files in %s. Total rules: %d", framework.Name, tmpDir, len(framework.Rules))
			}
		})
	}
}

// setupTestEnv creates files in the temp dir to satisfy the rule.
// Returns true if setup was successful, false if the rule is too complex to auto-mock.
func setupTestEnv(t *testing.T, dir string, framework *model.Framework) bool {
	// 如果没有规则，返回 false
	if len(framework.Rules) == 0 {
		return false
	}

	// 取第一条规则作为测试目标
	rule := framework.Rules[0]

	// 处理 Paths
	if len(rule.Paths) > 0 {
		// 对每条路径创建对应的文件或目录
		for _, pathPattern := range rule.Paths {
			filename := pathPattern

			// 如果模式以 / 结尾，它期望目录存在。
			// 在其中创建一个文件以便被索引。
			if strings.HasSuffix(pathPattern, "/") {
				filename = pathPattern + "index.php"
			} else if pathPattern == "*.go" {
				filename = "main.go"
			} else if pathPattern == "*.js" {
				filename = "index.js"
			} else if pathPattern == "*.json" {
				filename = "package.json"
			} else if pathPattern == "*.php" {
				filename = "index.php"
			} else if pathPattern == "*.py" {
				filename = "app.py"
			} else if pathPattern == "pom.xml" {
				filename = "pom.xml"
			} else if pathPattern == "go.mod" {
				filename = "go.mod"
			} else if strings.Contains(pathPattern, "*") {
				// Replace * with something concrete
				filename = strings.ReplaceAll(pathPattern, "*", "test_file")
			}

			// Create the file
			fullPath := filepath.Join(dir, filename)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Errorf("Failed to create dirs: %v", err)
				return false
			}

			// 如果是文件，写入空内容
			if !strings.HasSuffix(filename, "/") {
				if err := os.WriteFile(fullPath, []byte(""), 0644); err != nil {
					t.Errorf("Failed to write file: %v", err)
					return false
				}
			}
		}
	}

	// 处理 FileContents
	if len(rule.FileContents) > 0 {
		// 对每个文件创建对应的文件并写入内容
		for filePath, keywords := range rule.FileContents {
			// Create the file

			fullPath := filepath.Join(dir, strings.ReplaceAll(filePath, "*", "_"))
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Errorf("Failed to create dirs: %v", err)
				return false
			}

			// Construct content
			content := ""
			// Just concatenate all required strings
			for _, s := range keywords {
				content += s + "\n"
			}

			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Errorf("Failed to write file: %v", err)
				return false
			}
		}
	}

	return true
}
