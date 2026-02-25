package frameengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/winezer0/xcanvas/camodels"
)

// TestDetectFrameworks tests the framework detection functionality.
func TestDetectFrameworks(t *testing.T) {
	// We'll use the embedded rules directly without creating custom rules
	// This tests that embedded rules work correctly

	// Create a test project directory with Node.js backend files
	projectDir, err := os.MkdirTemp("", "test_project")
	if err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Create package.json file with Express dependency
	packageJsonContent := []byte(`{
  "name": "test-app",
  "dependencies": {
    "express": "^4.18.0"
  }
}`)

	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), packageJsonContent, 0644); err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create a new rule engine with embedded rules only (empty rules directory)
	ruleEngine, err := NewCanvasEngine("")
	if err != nil {
		t.Fatalf("Failed to create rule engine: %v", err)
	}

	// Build file index
	index, err := buildTestIndex(projectDir)
	if err != nil {
		t.Fatalf("Failed to build file index: %v", err)
	}

	// Detect frameworks in the test project
	result, err := ruleEngine.DetectFrameworks(index, []string{"JavaScript"})
	if err != nil {
		t.Fatalf("Failed to detect frameworks: %v", err)
	}

	// Check that Express was detected as a backend framework
	// There might be multiple frameworks detected, so we'll check if Express is among them
	expressFound := false
	for _, fw := range result.Frameworks {
		if fw.Name == "Express" {
			expressFound = true
			if fw.Language != "JavaScript" {
				t.Errorf("Expected Express framework language 'JavaScript', got '%s'", fw.Language)
			}
			if fw.Category != camodels.CategoryBackend {
				t.Errorf("Expected Express framework category 'backend', got '%s'", fw.Category)
			}
			break
		}
	}

	if !expressFound {
		t.Errorf("Expected Express framework to be detected")
	}
}

// TestDetectComponents tests the component detection functionality.
func TestDetectComponents(t *testing.T) {
	// We'll use the embedded rules directly without creating custom rules
	// This tests that embedded component rules work correctly

	// Create a test project directory with component files
	projectDir, err := os.MkdirTemp("", "test_project")
	if err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Create package.json file with lodash dependency
	packageJsonContent := []byte(`{
  "name": "test-app",
  "dependencies": {
    "lodash": "^4.17.0"
  }
}`)

	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), packageJsonContent, 0644); err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create a new rule engine with embedded rules only (empty rules directory)
	ruleEngine, err := NewCanvasEngine("")
	if err != nil {
		t.Fatalf("Failed to create rule engine: %v", err)
	}

	// Build file index
	index, err := buildTestIndex(projectDir)
	if err != nil {
		t.Fatalf("Failed to build file index: %v", err)
	}

	// Detect components in the test project
	result, err := ruleEngine.DetectFrameworks(index, []string{"JavaScript"})
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Check that lodash was detected as a frontend component (since we categorized JS libs as frontend by default if not backend specific)
	// There might be multiple components detected, so we'll check if lodash is among them
	lodashFound := false
	for _, comp := range result.Components {
		if comp.Name == "lodash" {
			lodashFound = true
			if comp.Language != "JavaScript" {
				t.Errorf("Expected lodash component language 'JavaScript', got '%s'", comp.Language)
			}
			if comp.Category != camodels.CategoryFrontend {
				t.Errorf("Expected lodash component category 'frontend', got '%s'", comp.Category)
			}
			break
		}
	}

	if !lodashFound {
		t.Errorf("Expected lodash component to be detected")
	}
}

// TestDetectVersion tests the version detection functionality.
func TestDetectVersion(t *testing.T) {
	// Create a temporary directory with test rules
	tempDir, err := os.MkdirTemp("", "test_version_rules")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write a test rule with version extraction
	customYamlContent := []byte(`- name: TestFramework
  type: framework
  language: JavaScript
  category: backend
  rules:
    - paths:
        - package.json
      file_contents:
        package.json:
          - "test-framework"
  version:
    - file_pattern: package.json
      patterns:
        - '"test-framework"\s*:\s*"([0-9]+\.[0-9]+(?:\.[0-9]+)?)"'
`)

	if err := os.WriteFile(filepath.Join(tempDir, "test.yml"), customYamlContent, 0644); err != nil {
		t.Fatalf("Failed to write custom rule file: %v", err)
	}

	// Create a test project directory
	projectDir, err := os.MkdirTemp("", "test_version_project")
	if err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}
	defer os.RemoveAll(projectDir)

	// Create package.json with version information
	packageJsonContent := []byte(`{
  "name": "test-app",
  "dependencies": {
    "test-framework": "1.2.3"
  }
}`)

	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), packageJsonContent, 0644); err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create a rule engine with the custom rules
	ruleEngine, err := NewCanvasEngine(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rule engine: %v", err)
	}

	// Build file index
	index, err := buildTestIndex(projectDir)
	if err != nil {
		t.Fatalf("Failed to build file index: %v", err)
	}

	// Detect frameworks in the test project
	result, err := ruleEngine.DetectFrameworks(index, []string{"JavaScript"})
	if err != nil {
		t.Fatalf("Failed to detect frameworks: %v", err)
	}

	// Check that TestFramework was detected with version
	testFrameworkFound := false
	for _, fw := range result.Frameworks {
		if fw.Name == "TestFramework" {
			testFrameworkFound = true
			if fw.Version != "1.2.3" {
				t.Errorf("Expected TestFramework version '1.2.3', got '%s'", fw.Version)
			}
			break
		}
	}

	if !testFrameworkFound {
		t.Errorf("Expected TestFramework to be detected")
	}
}

// buildTestIndex creates a file index for testing
func buildTestIndex(rootDir string) (*camodels.FileIndex, error) {
	index := camodels.NewFileIndex(rootDir)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(rootDir, path)
			relPath = filepath.ToSlash(relPath)
			fileName := filepath.Base(path)
			ext := filepath.Ext(path)
			index.AddFile(relPath, fileName, ext)
		}
		return nil
	})
	return index, err
}
