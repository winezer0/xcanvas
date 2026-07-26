package canvas

import (
	"context"
	"fmt"
	"time"

	"github.com/winezer0/xutils/logging"

	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xcanvas/internal/analyzer"
	"github.com/winezer0/xcanvas/internal/frameengine"
)

// Analyze performs a full analysis and returns a CanvasReport.
// This is the backward-compatible entry point without cancellation support.
func Analyze(path string, rulesDir string) (*camodels.CanvasReport, error) {
	return AnalyzeWithContext(context.Background(), path, rulesDir, DefaultOptions())
}

// Options configures resource limits for xcanvas analysis.
// Mirrors analyzer.WalkOptions for external consumers.
type Options struct {
	MaxFiles       int
	MaxFileSize    int64
	MaxDepth       int
	FollowSymlinks bool
}

// DefaultOptions returns production-safe defaults.
func DefaultOptions() Options {
	d := analyzer.DefaultWalkOptions()
	return Options{
		MaxFiles:       d.MaxFiles,
		MaxFileSize:    d.MaxFileSize,
		MaxDepth:       d.MaxDepth,
		FollowSymlinks: d.FollowSymlinks,
	}
}

// toWalkOptions converts public Options to internal analyzer.WalkOptions.
func (o Options) toWalkOptions() analyzer.WalkOptions {
	return analyzer.WalkOptions{
		MaxFiles:       o.MaxFiles,
		MaxFileSize:    o.MaxFileSize,
		MaxDepth:       o.MaxDepth,
		FollowSymlinks: o.FollowSymlinks,
	}
}

// AnalyzeWithContext performs a full analysis with context cancellation and resource limits.
func AnalyzeWithContext(ctx context.Context, path string, rulesDir string, opts Options) (*camodels.CanvasReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("xcanvas: canceled before analysis: %w", err)
	}

	// Initialize framework detection rule engine.
	canvasEngine, initErr := frameengine.NewCanvasEngine(rulesDir)
	if initErr != nil {
		return nil, fmt.Errorf("init canvas engine rules error: %v", initErr)
	}

	// Analyze code structure with context and resource limits.
	codeAnalyzer := analyzer.NewCodeAnalyzer()
	codeProfile, fileIndex, diag, analyzerErr := codeAnalyzer.AnalyzeCodeProfileWithContext(ctx, path, opts.toWalkOptions())
	if analyzerErr != nil {
		return nil, fmt.Errorf("error analyzing code profile: %w", analyzerErr)
	}
	if diag != nil && diag.Truncated {
		logging.Warnf("xcanvas: file traversal truncated at %d files", opts.toWalkOptions().Normalize().MaxFiles)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("xcanvas: canceled after profile analysis: %w", err)
	}

	// Detect frameworks and components using rules.
	detectInfo, detectErr := canvasEngine.DetectFrameworks(fileIndex, codeProfile.Expands)
	if detectErr != nil {
		return nil, fmt.Errorf("error detecting frameworks and components: %v", detectErr)
	}

	report := &camodels.CanvasReport{
		CodeProfile: *codeProfile,
		Detection:   *detectInfo,
		Timestamp:   time.Now(),
	}
	return report, nil
}

// AnalyzeProjectInfoWithCanvas 初始化項目信息 并分析canvasReport
func AnalyzeProjectInfoWithCanvas(projectName, projectPath, canvasRulesDir string) *camodels.ProjectInfo {
	// 初始化项目画像信息
	projectInfo := camodels.NewEmptyProjectInfo(projectName, projectPath)
	// 获取 xcanvas 代码画像 使用 Analyze 获取语言、框架和组件信息
	canvasReport, err := Analyze(projectPath, canvasRulesDir)
	if err != nil {
		logging.Errorf("detection canvas info error: %v", err)
		return projectInfo
	}

	// 补充canvas信息
	simpleCanvas := canvasReport.ToSimpleReport()
	projectInfo.FilesCount = simpleCanvas.TotalFiles
	projectInfo.Languages = simpleCanvas.Languages
	projectInfo.Frameworks = simpleCanvas.Frameworks
	projectInfo.Components = simpleCanvas.Components
	projectInfo.BackendLanguages = simpleCanvas.BackendLanguages
	projectInfo.FrontendLanguages = simpleCanvas.FrontendLanguages
	return projectInfo
}
