package canvas

import (
	"fmt"
	"time"

	"github.com/winezer0/xcanvas/internal/model"
	"github.com/winezer0/xcanvas/internal/utils"
	"github.com/winezer0/xcanvas/models"
)

// PrintReport outputs the analysis report in text format.
func PrintReport(report *models.CanvasReport) {
	fmt.Println("CodeCanvas Analysis Report")
	fmt.Println("=========================")
	fmt.Printf("Path: %s\n", report.CodeProfile.Path)
	fmt.Printf("Total Files: %d\n", report.CodeProfile.TotalFiles)
	fmt.Printf("Total Lines: %d\n", report.CodeProfile.TotalLines)
	fmt.Println()

	// Frontend languages
	if len(report.CodeProfile.FrontendLanguages) > 0 {
		fmt.Println("Frontend LanguageInfos:")
		for _, lang := range report.CodeProfile.FrontendLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// Backend languages
	if len(report.CodeProfile.BackendLanguages) > 0 {
		fmt.Println("Backend LanguageInfos:")
		for _, lang := range report.CodeProfile.BackendLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// Desktop languages
	if len(report.CodeProfile.DesktopLanguages) > 0 {
		fmt.Println("Desktop LanguageInfos:")
		for _, lang := range report.CodeProfile.DesktopLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// Other languages
	if len(report.CodeProfile.OtherLanguages) > 0 {
		fmt.Println("Other LanguageInfos:")
		for _, lang := range report.CodeProfile.OtherLanguages {
			fmt.Printf("- %s\n", lang)
		}
		fmt.Println()
	}

	// All languages (verbose only)
	fmt.Println("All LanguageInfos:")
	for _, lang := range report.CodeProfile.LanguageInfos {
		fmt.Printf("- %s: %d files, %d lines\n", lang.Name, lang.Files, lang.CodeLines)
	}
	fmt.Println()

	// Frameworks
	if len(report.Detection.Frameworks) > 0 {
		PrintDetectedItems("Detected Frameworks", report.Detection.Frameworks)
	} else {
		fmt.Printf("Detected Frameworks Is Empty !!!\n")
	}

	// Components
	if len(report.Detection.Components) > 0 {
		PrintDetectedItems("Detected Components", report.Detection.Components)
	} else {
		fmt.Printf("Detected Components Is Empty !!!\n")
	}

	fmt.Printf("Generated: %s\n", report.Timestamp.Format(time.RFC1123))

	simpleReport := ToSimpleReport(report)

	fmt.Printf("Simple Report:\n%s", utils.ToJson(simpleReport))
}

func PrintDetectedItems(title string, items []model.DetectedItem) {
	fmt.Println(title + ":")

	// Group by category
	byCategory := make(map[string][]model.DetectedItem)
	for _, item := range items {
		byCategory[item.Category] = append(byCategory[item.Category], item)
	}

	categories := model.AllCategory
	for _, cat := range categories {
		PrintCategoryItems(cat, byCategory[cat])
	}

	// Handle any other categories
	for cat, items := range byCategory {
		isKnown := false
		for _, known := range categories {
			if cat == known {
				isKnown = true
				break
			}
		}
		if !isKnown && len(items) > 0 {
			PrintCategoryItems(cat, items)
		}
	}
	fmt.Println()
}

func PrintCategoryItems(category string, items []model.DetectedItem) {
	if len(items) > 0 {
		fmt.Printf("  [%s]\n", category)
		for _, item := range items {
			fmt.Printf("  - %s (%s)\n", item.Name, item.Language)
			if item.Version != "" {
				fmt.Printf("    Version: %s\n", item.Version)
			}
			if item.Evidence != "" {
				fmt.Printf("    Evidence: %s\n", item.Evidence)
			}
		}
	}
}
