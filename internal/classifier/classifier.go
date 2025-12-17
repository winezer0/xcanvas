package classifier

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/winezer0/codecanvas/internal/model"
)

type LanguageSyntaxFeatures struct {
	Name         string   `json:"name"`
	Tokens       []string `json:"tokens"`
	FilePatterns []string `json:"file_patterns"`
	Dependencies []string `json:"dependencies"`
}

type LanguageClassificationRule struct {
	Name     string                 `json:"name"`
	Category string                 `json:"category"`
	Features LanguageSyntaxFeatures `json:"features"`
}

type LanguageClassifier struct {
	rules map[string]LanguageClassificationRule
}

func NewLanguageClassifier() *LanguageClassifier {
	c := &LanguageClassifier{rules: map[string]LanguageClassificationRule{}}
	c.bootstrap()
	return c
}

func (c *LanguageClassifier) LoadFromFile(path string) error {
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var items []LanguageClassificationRule
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	for _, it := range items {
		c.rules[strings.ToLower(it.Name)] = it
	}
	return nil
}

func (c *LanguageClassifier) DetectCategories(root string, langs []model.LanguageInfo) (frontend []string, backend []string, desktop []string, all []string) {
	fset := make(map[string]bool)
	bset := make(map[string]bool)
	dset := make(map[string]bool)
	allSet := make(map[string]bool) // 用于去重所有语言

	deps := c.readPackageJSONDeps(root)

	for _, li := range langs {
		name := strings.ToLower(li.Name)
		r, ok := c.rules[name]
		if ok {
			cat := c.applyHeuristics(root, r, deps)
			switch cat {
			case model.CategoryFrontend:
				fset[li.Name] = true
			case model.CategoryBackend:
				bset[li.Name] = true
			case model.CategoryDesktop:
				dset[li.Name] = true
			}
		} else {
			if model.FrontendLanguageSet[li.Name] {
				fset[li.Name] = true
			} else if model.BackendLanguageSet[li.Name] {
				bset[li.Name] = true
			}
			// 注意：Desktop 只能通过 rules 判断？根据原逻辑，是的。
		}

		// 所有语言都加入 allSet（去重）
		allSet[li.Name] = true
	}

	// 提取结果（保持顺序无关，若需排序可加 sort.Strings）
	frontend = keys(fset)
	backend = keys(bset)
	desktop = keys(dset)
	all = keys(allSet)

	return
}

// keys 辅助函数：从 map[string]bool 提取所有 key
func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func (c *LanguageClassifier) applyHeuristics(root string, r LanguageClassificationRule, deps map[string]bool) string {
	name := strings.ToLower(r.Name)
	if name == "javascript" || name == "typescript" {
		if deps["express"] || deps["koa"] || deps["nestjs"] || deps["fastify"] || deps["hapi"] {
			return model.CategoryBackend
		}
		if deps["react"] || deps["vue"] || deps["@angular/core"] || deps["next"] || deps["nuxt"] {
			return model.CategoryFrontend
		}
	}
	if r.Category == model.CategoryFrontend || r.Category == model.CategoryBackend {
		return r.Category
	}
	if len(r.Features.Dependencies) > 0 {
		for _, d := range r.Features.Dependencies {
			if deps[strings.ToLower(d)] {
				return model.CategoryFrontend
			}
		}
	}
	for _, p := range r.Features.FilePatterns {
		matches, _ := filepath.Glob(filepath.Join(root, p))
		if len(matches) > 0 {
			if strings.Contains(strings.ToLower(r.Name), "jsx") ||
				strings.Contains(strings.ToLower(r.Name), "tsx") ||
				strings.Contains(strings.ToLower(r.Name), "html") ||
				strings.Contains(strings.ToLower(r.Name), "css") {
				return model.CategoryFrontend
			}
			return model.CategoryBackend
		}
	}
	return "other"
}

func (c *LanguageClassifier) readPackageJSONDeps(root string) map[string]bool {
	res := map[string]bool{}
	p := filepath.Join(root, "package.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return res
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	for _, key := range []string{"dependencies", "devDependencies", "peerDependencies"} {
		v, ok := m[key]
		if ok {
			mm, _ := v.(map[string]any)
			for k := range mm {
				res[strings.ToLower(k)] = true
			}
		}
	}
	return res
}

func (c *LanguageClassifier) bootstrap() {
	presets := []LanguageClassificationRule{
		{Name: "JavaScript", Category: model.CategoryFrontend, Features: LanguageSyntaxFeatures{Dependencies: []string{"react", "vue", "@angular/core"}, FilePatterns: []string{"**/*.jsx", "**/*.tsx"}}},
		{Name: "TypeScript", Category: model.CategoryFrontend, Features: LanguageSyntaxFeatures{Dependencies: []string{"react", "vue", "@angular/core"}, FilePatterns: []string{"**/*.tsx"}}},
		{Name: "JSX", Category: model.CategoryFrontend},
		{Name: "TSX", Category: model.CategoryFrontend},
		{Name: "HTML", Category: model.CategoryFrontend},
		{Name: "CSS", Category: model.CategoryFrontend},
		{Name: "SCSS", Category: model.CategoryFrontend},
		{Name: "Less", Category: model.CategoryFrontend},
		{Name: "WebAssembly", Category: model.CategoryFrontend, Features: LanguageSyntaxFeatures{FilePatterns: []string{"**/*.wasm"}}},
		{Name: "Java", Category: model.CategoryBackend},
		{Name: "Kotlin", Category: model.CategoryBackend},
		{Name: "Python", Category: model.CategoryBackend},
		{Name: "Go", Category: model.CategoryBackend},
		{Name: "Ruby", Category: model.CategoryBackend},
		{Name: "PHP", Category: model.CategoryBackend},
		{Name: "C#", Category: model.CategoryBackend},
		{Name: ".NET", Category: model.CategoryBackend},
		{Name: "Rust", Category: model.CategoryBackend},
		{Name: "Node.js", Category: model.CategoryBackend, Features: LanguageSyntaxFeatures{Dependencies: []string{"express", "koa", "nestjs"}, FilePatterns: []string{"server.js", "server.ts", "src/server/*"}}},
	}
	for _, it := range presets {
		c.rules[strings.ToLower(it.Name)] = it
	}
}
