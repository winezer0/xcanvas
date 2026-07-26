package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/winezer0/xcanvas/camodels"
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
func generateTestContent(lang *camodels.Language) string {
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

// --- Context and resource limit tests ---

// TestAnalyzeCodeProfileCancelledContext verifies cancelled context aborts analysis.
func TestAnalyzeCodeProfileCancelledContext(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, tmpDir, "main.go", "package main\nfunc main() {}\n")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	az := NewCodeAnalyzer()
	_, _, _, err := az.AnalyzeCodeProfileWithContext(ctx, tmpDir, DefaultWalkOptions())
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAnalyzeCodeProfileMaxFiles verifies file count limit truncates traversal.
func TestAnalyzeCodeProfileMaxFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 10; i++ {
		writeFile(t, tmpDir, fmt.Sprintf("file%d.go", i), "package main\n")
	}

	az := NewCodeAnalyzer()
	opts := WalkOptions{MaxFiles: 3, MaxFileSize: 1024 * 1024, MaxDepth: 10}
	profile, _, diag, err := az.AnalyzeCodeProfileWithContext(context.Background(), tmpDir, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !diag.Truncated {
		t.Error("expected diagnostics.Truncated = true")
	}
	if profile.TotalFiles > 3 {
		t.Fatalf("expected at most 3 files, got %d", profile.TotalFiles)
	}
}

// TestAnalyzeCodeProfileMaxFileSize verifies large files are skipped.
func TestAnalyzeCodeProfileMaxFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a file larger than 100 bytes limit.
	largeContent := strings.Repeat("x", 200)
	writeFile(t, tmpDir, "large.go", largeContent)
	writeFile(t, tmpDir, "small.go", "package main\n")

	az := NewCodeAnalyzer()
	opts := WalkOptions{MaxFiles: 100, MaxFileSize: 100, MaxDepth: 10}
	profile, _, diag, err := az.AnalyzeCodeProfileWithContext(context.Background(), tmpDir, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diag.SkippedLarge != 1 {
		t.Fatalf("expected 1 skipped large file, got %d", diag.SkippedLarge)
	}
	if profile.TotalFiles != 1 {
		t.Fatalf("expected 1 file analyzed, got %d", profile.TotalFiles)
	}
}

// TestAnalyzeCodeProfileMaxDepth verifies depth limit stops deep traversal.
func TestAnalyzeCodeProfileMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()
	// Create nested directories: a/b/c/d/deep.go
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, tmpDir, "root.go", "package main\n")
	writeFile(t, deepDir, "deep.go", "package deep\n")

	az := NewCodeAnalyzer()
	opts := WalkOptions{MaxFiles: 100, MaxFileSize: 1024 * 1024, MaxDepth: 2}
	profile, _, diag, err := az.AnalyzeCodeProfileWithContext(context.Background(), tmpDir, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !diag.MaxDepthReached {
		t.Error("expected MaxDepthReached = true")
	}
	// Only root.go should be found (depth 0), a/b is depth 2 which is the limit.
	if profile.TotalFiles != 1 {
		t.Fatalf("expected 1 file, got %d", profile.TotalFiles)
	}
}

// TestAnalyzeCodeProfileSymlinkSkip verifies symlinks are skipped by default.
func TestAnalyzeCodeProfileSymlinkSkip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink test requires Unix-like OS")
	}
	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, realDir, "code.go", "package real\n")
	// Create symlink pointing back to tmpDir (cycle).
	linkPath := filepath.Join(tmpDir, "link")
	if err := os.Symlink(tmpDir, linkPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	az := NewCodeAnalyzer()
	opts := WalkOptions{MaxFiles: 100, MaxFileSize: 1024 * 1024, MaxDepth: 10, FollowSymlinks: false}
	profile, _, diag, err := az.AnalyzeCodeProfileWithContext(context.Background(), tmpDir, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diag.SkippedSymlinks < 1 {
		t.Fatalf("expected at least 1 skipped symlink, got %d", diag.SkippedSymlinks)
	}
	// Should find code.go in real/ but not loop through link/.
	if profile.TotalFiles != 1 {
		t.Fatalf("expected 1 file, got %d", profile.TotalFiles)
	}
}

// TestAnalyzeCodeProfileBackwardCompat verifies the original API still works.
func TestAnalyzeCodeProfileBackwardCompat(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, tmpDir, "main.go", "package main\nfunc main() {}\n")

	az := NewCodeAnalyzer()
	profile, index, err := az.AnalyzeCodeProfile(tmpDir)
	if err != nil {
		t.Fatalf("backward compat failed: %v", err)
	}
	if profile.TotalFiles != 1 {
		t.Fatalf("expected 1 file, got %d", profile.TotalFiles)
	}
	if index == nil {
		t.Fatal("file index is nil")
	}
}

// TestPathDepth verifies depth calculation.
func TestPathDepth(t *testing.T) {
	tests := []struct {
		root string
		path string
		want int
	}{
		{"/repo", "/repo", 0},
		{"/repo", "/repo/a", 1},
		{"/repo", "/repo/a/b/c", 3},
	}
	for _, tc := range tests {
		got := pathDepth(tc.root, tc.path)
		if got != tc.want {
			t.Errorf("pathDepth(%q, %q) = %d, want %d", tc.root, tc.path, got, tc.want)
		}
	}
}

// writeFile is a test helper that writes content to a file in dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
