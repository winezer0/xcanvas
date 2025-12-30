package langengine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/winezer0/xcanvas/internal/model"
)

func TestDetectCategories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "classifier_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := NewLangClassifier()

	tests := []struct {
		name         string
		languages    []model.LangInfo
		packageJson  map[string]any
		wantFrontend []string
		wantBackend  []string
	}{
		{
			name: "Basic Backend",
			languages: []model.LangInfo{
				{Name: "Go"},
				{Name: "Java"},
			},
			wantBackend: []string{"Go", "Java"},
		},
		{
			name: "Basic Frontend",
			languages: []model.LangInfo{
				{Name: "HTML"},
				{Name: "CSS"},
			},
			wantFrontend: []string{"HTML", "CSS"},
		},
		{
			name: "JS with React (Frontend)",
			languages: []model.LangInfo{
				{Name: "JavaScript"},
			},
			packageJson: map[string]any{
				"dependencies": map[string]any{
					"react": "^17.0.0",
				},
			},
			wantFrontend: []string{"JavaScript"},
		},
		{
			name: "Mixed Project",
			languages: []model.LangInfo{
				{Name: "Go"},
				{Name: "TypeScript"},
			},
			packageJson: map[string]any{
				"dependencies": map[string]any{
					"react": "^17.0.0",
				},
			},
			wantFrontend: []string{"TypeScript"},
			wantBackend:  []string{"Go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create package.json if needed
			if tt.packageJson != nil {
				b, _ := json.Marshal(tt.packageJson)
				os.WriteFile(filepath.Join(tmpDir, "package.json"), b, 0644)
			} else {
				os.Remove(filepath.Join(tmpDir, "package.json"))
			}

			frontend, backend, _, _, _, _ := c.DetectCategories(tmpDir, tt.languages)

			checkList(t, "Frontend", frontend, tt.wantFrontend)
			checkList(t, "Backend", backend, tt.wantBackend)
		})
	}
}

func checkList(t *testing.T, cat string, got, want []string) {
	if len(got) != len(want) {
		t.Errorf("%s: got %v, want %v", cat, got, want)
		return
	}
	seen := make(map[string]bool)
	for _, s := range got {
		seen[s] = true
	}
	for _, s := range want {
		if !seen[s] {
			t.Errorf("%s: missing %s", cat, s)
		}
	}
}
