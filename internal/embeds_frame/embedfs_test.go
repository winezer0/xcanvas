package embeds_frame

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
	files, err := fs.Glob(FrameEmbedFS, "*.yml")
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
			content, err := FrameEmbedFS.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", file, err)
			}

			// Try to parse as a slice of frameworks
			var frameworks []struct {
				Name     string `yaml:"name"`
				Type     string `yaml:"type"`
				Language string `yaml:"language"`
				Category string `yaml:"category"`
				Rules    []struct {
					Paths        []string            `yaml:"paths"`
					FileContents map[string][]string `yaml:"file_contents"`
				} `yaml:"rules"`
				// Version extraction rules at the framework level
				Versions []struct {
					FilePattern string   `yaml:"file_pattern"`
					Patterns    []string `yaml:"patterns"`
				} `yaml:"version"`
			}

			if err := yaml.Unmarshal(content, &frameworks); err != nil {
				// If slice parsing fails, try as multi-document YAML
				decoder := yaml.NewDecoder(strings.NewReader(string(content)))
				for {
					var framework struct {
						Name     string `yaml:"name"`
						Type     string `yaml:"type"`
						Language string `yaml:"language"`
						Category string `yaml:"category"`
						Rules    []struct {
							Paths        []string            `yaml:"paths"`
							FileContents map[string][]string `yaml:"file_contents"`
						} `yaml:"rules"`
						// Version extraction rules at the framework level
						Versions []struct {
							FilePattern string   `yaml:"file_pattern"`
							Patterns    []string `yaml:"patterns"`
						} `yaml:"version"`
					}

					if err := decoder.Decode(&framework); err != nil {
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
