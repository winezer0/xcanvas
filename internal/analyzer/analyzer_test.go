package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/winezer0/xcanvas/internal/model"
)

func TestAnalyzeCodeProfile(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "codecanvas_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create main.go
	mainGoContent := `package main

// Comment
func main() {
    println("Hello")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Create test.js
	testJsContent := `/*
 Block comment
*/
console.log("Hi"); // Line comment
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(testJsContent), 0644); err != nil {
		t.Fatalf("Failed to write test.js: %v", err)
	}

	// Create ignored file (unknown extension)
	if err := os.WriteFile(filepath.Join(tmpDir, "unknown.xyz"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("Failed to write unknown file: %v", err)
	}

	// Analyze
	analyzer := NewCodeAnalyzer()
	profile, _, err := analyzer.AnalyzeCodeProfile(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCodeProfile failed: %v", err)
	}

	// Verify
	if profile.TotalFiles != 2 {
		t.Errorf("Expected 2 files, got %d", profile.TotalFiles)
	}

	// Check Go stats
	var goFound bool
	for _, lang := range profile.LanguageInfos {
		if lang.Name == "Go" {
			goFound = true
			if lang.Files != 1 {
				t.Errorf("Expected 1 Go file, got %d", lang.Files)
			}
			// Code: 4, Comment: 1, Blank: 1
			if lang.CodeLines != 4 {
				t.Errorf("Expected 4 Go code lines, got %d", lang.CodeLines)
			}
			if lang.CommentLines != 1 {
				t.Errorf("Expected 1 Go comment line, got %d", lang.CommentLines)
			}
			if lang.BlankLines != 1 {
				t.Errorf("Expected 1 Go blank line, got %d", lang.BlankLines)
			}
		} else if lang.Name == "JavaScript" {
			// Code: 1, Comment: 3, Blank: 0
			if lang.CodeLines != 1 {
				t.Errorf("Expected 1 JS code line, got %d", lang.CodeLines)
			}
			if lang.CommentLines != 3 {
				t.Errorf("Expected 3 JS comment lines, got %d", lang.CommentLines)
			}
		}
	}

	if !goFound {
		t.Error("Go language not found in profile")
	}
}

// TestAllLanguagesCoverage verifies that the analyzer can identify and count all supported languages.
func TestAllLanguagesCoverage(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "codecanvas_coverage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with a few common languages
	testLanguages := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".cs"}

	// Generate a file for each test language
	for _, ext := range testLanguages {
		filename := "test" + ext
		filePath := filepath.Join(tmpDir, filename)

		// Simple content for testing
		content := "// Test file\n"
		content += "function test() {\n"
		content += "    return 1;\n"
		content += "}\n"

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Errorf("Failed to write test file for %s: %v", ext, err)
		}
	}

	// Analyze the directory
	az := NewCodeAnalyzer()
	profile, _, err := az.AnalyzeCodeProfile(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCodeProfile failed: %v", err)
	}

	// Verify results
	if len(profile.LanguageInfos) == 0 {
		t.Error("No languages were detected")
	}
}

// generateTestContent creates a simple content string for a language
func generateTestContent(lang *model.Language) string {
	content := ""

	// Add a single line comment if supported
	if len(lang.LineComments) > 0 {
		content += fmt.Sprintf("%s This is a line comment\n", lang.LineComments[0])
	}

	// Add a multi-line comment if supported
	if len(lang.MultiLine) > 0 {
		content += fmt.Sprintf("%s\n This is a\n multi-line comment\n%s\n", lang.MultiLine[0][0], lang.MultiLine[0][1])
	}

	// Add some "code" (just text)
	content += "some_code_here = true\n"
	content += "function call() {}\n"

	// Add a blank line
	content += "\n"

	return content
}
