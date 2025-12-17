package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestAllLanguagesCoverage verifies that the analyzer can identify and count all supported languages.
func TestAllLanguagesCoverage(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "codecanvas_coverage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate a file for each supported language
	for _, lang := range languages {
		filename := ""
		if len(lang.Extensions) > 0 {
			filename = "test" + lang.Extensions[0]
		} else if len(lang.Filenames) > 0 {
			filename = lang.Filenames[0]
		} else {
			t.Errorf("Language %s has no extensions or filenames defined", lang.Name)
			continue
		}

		content := generateTestContent(lang)
		filePath := filepath.Join(tmpDir, filename)

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Errorf("Failed to write test file for %s: %v", lang.Name, err)
		}
	}

	// Analyze the directory
	az := NewCodeAnalyzer()
	profile, _, err := az.AnalyzeCodeProfile(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCodeProfile failed: %v", err)
	}

	// Verify results
	foundLanguages := make(map[string]bool)
	for _, l := range profile.LanguageInfos {
		foundLanguages[l.Name] = true
		// Verify counts (expecting roughly 1 file, >0 code lines per language)
		if l.Files != 1 {
			t.Errorf("Language %s: expected 1 file, got %d", l.Name, l.Files)
		}
		if l.CodeLines == 0 && l.CommentLines == 0 && l.BlankLines == 0 {
			t.Errorf("Language %s: stats are all zero", l.Name)
		}
	}

	// Check if all supported languages were found
	for _, lang := range languages {
		if !foundLanguages[lang.Name] {
			t.Errorf("Language %s was not detected", lang.Name)
		}
	}
}

// generateTestContent creates a simple content string for a language
func generateTestContent(lang LanguageDefinition) string {
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
