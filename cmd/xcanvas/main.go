// Package main provides the CLI interface for CodeCanvas.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/winezer0/codecanvas/canvas"
	"github.com/winezer0/codecanvas/internal/model"

	"github.com/jessevdk/go-flags"
)

// Options defines the command-line parameters for CodeCanvas.
type Options struct {
	// Analysis parameters
	Path     string `short:"p" long:"path" description:"Path to the codebase to analyze"`
	RulesDir string `short:"r" long:"rules" description:"Directory containing detection RulesDirDir" default:"./rules"`
	Output   string `short:"o" long:"output" description:"Write JSON to path or URL"`
}

const (
	Version = "0.1.0"
)

func main() {
	// 解析命令行参数
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)

	// Add help description
	parser.LongDescription = `
CodeCanvas - A lightweight, dependency-free, embeddable code profiling engine for DevSecOps.

CodeCanvas provides standardized project technology stack identification, including:
- Multi-dimensional software component classification
- Framework and component detection
- Language composition analysis

Usage:
  codecanvas --path <directory> [options]
  codecanvas --list 
`

	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
	}
	// 进行路径分析
	if opts.Path != "" {
		// Analyze operation
		handleAnalyzeCommand(&opts)
	}
}

// handleAnalyzeCommand handles the "analyze" command.
func handleAnalyzeCommand(cmd *Options) {
	// Use the facade to perform analysis
	report, err := canvas.Analyze(cmd.Path, cmd.RulesDir)
	if err != nil {
		fmt.Printf("Error analyzing code profile: %v\n", err)
		os.Exit(1)
	}

	// 输出Json结果
	if cmd.Output != "" {
		if err := WriteJSON(cmd.Output, report); err != nil {
			fmt.Printf("Error writing output: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 输出命令行报告
	outputText(report)
}

// outputText outputs the analysis report in text format.
func outputText(report *model.CanvasReport) {
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

	// All languages (verbose only)
	fmt.Println("All LanguageInfos:")
	for _, lang := range report.CodeProfile.LanguageInfos {
		fmt.Printf("- %s: %d files, %d lines\n", lang.Name, lang.Files, lang.CodeLines)
	}
	fmt.Println()

	// Frameworks
	if len(report.Detection.Frameworks) > 0 {
		printDetectedItems("Detected Frameworks", report.Detection.Frameworks)
	}

	// Components
	if len(report.Detection.Components) > 0 {
		printDetectedItems("Detected Components", report.Detection.Components)
	}

	fmt.Printf("Generated: %s\n", report.Timestamp.Format(time.RFC1123))
}

func printDetectedItems(title string, items []model.DetectedItem) {
	fmt.Println(title + ":")

	// Group by category
	byCategory := make(map[string][]model.DetectedItem)
	for _, item := range items {
		byCategory[item.Category] = append(byCategory[item.Category], item)
	}

	categories := []string{model.CategoryFrontend, model.CategoryBackend, model.CategoryDesktop}
	for _, cat := range categories {
		printCategoryItems(cat, byCategory[cat])
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
			printCategoryItems(cat, items)
		}
	}
	fmt.Println()
}

func printCategoryItems(category string, items []model.DetectedItem) {
	if len(items) > 0 {
		fmt.Printf("  [%s]\n", category)
		for _, item := range items {
			fmt.Printf("  - %s (%s)\n", item.Name, item.Language)
			if item.Version != "" {
				fmt.Printf("    Version: %s\n", item.Version)
			}
			if item.Confidence != "" {
				fmt.Printf("    Confidence: %s\n", item.Confidence)
			}
			if item.Evidence != "" {
				fmt.Printf("    Evidence: %s\n", item.Evidence)
			}
		}
	}
}

func WriteJSON(path string, data interface{}) error {
	payload, _ := json.MarshalIndent(data, "", "  ")
	if !filepath.IsAbs(path) {
		path = filepath.Clean(path)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0644)
}
