package analyzer

import (
	"os"
	"path/filepath"
	"testing"
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
