package frameengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/winezer0/xcanvas/internal/model"
)

// TestNewRuleEngine tests the creation of a new rule engine with rules loaded from a directory.
func TestNewRuleEngine(t *testing.T) {
	// Create a temporary directory with test rules
	tempDir, err := os.MkdirTemp("", "test_rules")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write a test rule that will override the embedded Spring Boot rule
	yamlContent := []byte(`- name: Spring Boot
  type: framework
  language: Java
  category: backend
  levels:
    L1:
      paths:
      - pom.xml
      contains:
      - spring-boot-starter
`)

	if err := os.WriteFile(filepath.Join(tempDir, "java.yml"), yamlContent, 0644); err != nil {
		t.Fatalf("Failed to write test rule file: %v", err)
	}

	// Create a new rule engine with the test rules
	ruleEngine, err := InitCanvasEngine(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rule engine: %v", err)
	}

	// Check that the rule engine has at least the embedded rules plus our custom rule
	frameworks := ruleEngine.GetSupportedFrameworks()
	if len(frameworks) < 5 { // We have at least 5 embedded frameworks
		t.Errorf("Expected at least 5 frameworks (embedded + custom), got %d", len(frameworks))
	}

	// Check that Spring Boot rule is present (it should be either the embedded one or our custom one)
	springBootFound := false
	for _, fw := range frameworks {
		if fw.Name == "Spring Boot" && fw.Language == "Java" {
			springBootFound = true
			break
		}
	}

	if !springBootFound {
		t.Errorf("Expected Spring Boot framework to be present")
	}
}

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
	ruleEngine, err := InitCanvasEngine("")
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
			if fw.Category != model.CategoryBackend {
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
	ruleEngine, err := InitCanvasEngine("")
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
			if comp.Category != model.CategoryFrontend {
				t.Errorf("Expected lodash component category 'frontend', got '%s'", comp.Category)
			}
			break
		}
	}

	if !lodashFound {
		t.Errorf("Expected lodash component to be detected")
	}
}

// TestEdgeCases tests edge cases in the rule engine functionality.
func TestEdgeCases(t *testing.T) {
	// Test with empty rules directory (should still load embedded rules)
	emptyDir, err := os.MkdirTemp("", "test_empty_rules")
	if err != nil {
		t.Fatalf("Failed to create empty temp directory: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	// Create a rule engine with empty rules directory
	ruleEngine, err := InitCanvasEngine(emptyDir)
	if err != nil {
		t.Fatalf("Failed to create rule engine with empty directory: %v", err)
	}

	// Check that embedded frameworks are still supported
	if frameworks := ruleEngine.GetSupportedFrameworks(); len(frameworks) == 0 {
		t.Errorf("Expected at least some embedded frameworks, got 0")
	}

	// Check that embedded components are still supported
	if components := ruleEngine.GetSupportedComponents(); len(components) == 0 {
		t.Errorf("Expected at least some embedded components, got 0")
	}

	// Test detection with no matching files
	// Build file index for current directory
	index, err := buildTestIndex(".")
	if err != nil {
		t.Fatalf("Failed to build file index: %v", err)
	}

	result, err := ruleEngine.DetectFrameworks(index, []string{"Java"})
	if err != nil {
		t.Fatalf("Failed to detect frameworks with no matching files: %v", err)
	}

	if len(result.Frameworks) != 0 {
		t.Errorf("Expected 0 frameworks detected, got %d", len(result.Frameworks))
	}

	if len(result.Components) != 0 {
		t.Errorf("Expected 0 components detected, got %d", len(result.Components))
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
	ruleEngine, err := InitCanvasEngine(tempDir)
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

// TestEmbeddedRulesLoad tests that embedded rules are correctly loaded
func TestEmbeddedRulesLoad(t *testing.T) {
	// Create a new rule engine without any custom rules (should use embedded only)
	ruleEngine, err := InitCanvasEngine("")
	if err != nil {
		t.Fatalf("Failed to create rule engine: %v", err)
	}

	// Get all supported frameworks
	frameworks := ruleEngine.GetSupportedFrameworks()
	if len(frameworks) == 0 {
		t.Error("Expected at least some embedded frameworks, got none")
	}

	// Check specific frameworks
	expectedFrameworks := []string{"Express", "React", "Vue.js", "Angular", "Spring Boot", "Django", "Gin", "Echo"}
	for _, expected := range expectedFrameworks {
		found := false
		for _, fw := range frameworks {
			if fw.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected framework '%s' to be detected in embedded rules", expected)
		}
	}

	// Get all supported components
	components := ruleEngine.GetSupportedComponents()
	if len(components) == 0 {
		t.Error("Expected at least some embedded components, got none")
	}

	// Check specific components
	expectedComponents := []string{"lodash", "axios", "requests"}
	for _, expected := range expectedComponents {
		found := false
		for _, comp := range components {
			if comp.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected component '%s' to be detected in embedded rules", expected)
		}
	}
}

// buildTestIndex creates a file index for testing
func buildTestIndex(rootDir string) (*model.FileIndex, error) {
	index := model.NewFileIndex(rootDir)
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
