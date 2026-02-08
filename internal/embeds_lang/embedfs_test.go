package embeds_lang

import (
	"io"
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestAllYamlFiles tests that all embedded YAML files are valid
func TestAllYamlFiles(t *testing.T) {
	// Get all yml files from the embedded filesystem
	files, err := fs.Glob(LanguageEmbedFS, "*.yml")
	if err != nil {
		t.Fatalf("Failed to get YAML files: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("No YAML files found in embedded filesystem")
	}

	// Test each YAML file
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			// Read file content
			content, err := LanguageEmbedFS.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", file, err)
			}

			// Try to parse as a slice of language rules
			var languages []struct {
				Name     string `yaml:"name"`
				Type     string `yaml:"type"`
				Language string `yaml:"language"`
				Category string `yaml:"category"`
				Rules    []struct {
					FilePatterns []string            `yaml:"file_patterns"`
					FileContents map[string][]string `yaml:"file_contents"`
				} `yaml:"rules"`
			}

			if err := yaml.Unmarshal(content, &languages); err != nil {
				// If slice parsing fails, try as multi-document YAML
				decoder := yaml.NewDecoder(strings.NewReader(string(content)))
				for {
					var language struct {
						Name     string `yaml:"name"`
						Type     string `yaml:"type"`
						Language string `yaml:"language"`
						Category string `yaml:"category"`
						Rules    []struct {
							FilePatterns []string            `yaml:"file_patterns"`
							FileContents map[string][]string `yaml:"file_contents"`
						} `yaml:"rules"`
					}

					if err := decoder.Decode(&language); err != nil {
						if err == io.EOF {
							break
						}
						t.Fatalf("Failed to parse YAML file %s: %v", file, err)
					}
				}
			}
		})
	}
}
