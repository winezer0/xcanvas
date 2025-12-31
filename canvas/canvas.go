package canvas

import (
	"context"
	"fmt"
	"time"

	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xcanvas/internal/analyzer"
	"github.com/winezer0/xcanvas/internal/frameengine"
)

// Analyze performs a full analysis and returns a CanvasReport.
func Analyze(path string, rulesDir string) (*camodels.CanvasReport, error) {
	ctx := context.Background()

	// Analyze code profile
	az := analyzer.NewCodeAnalyzer()
	profile, index, err := az.AnalyzeCodeProfile(path)
	if err != nil {
		return nil, fmt.Errorf("error analyzing code profile: %v", err)
	}

	// Create rule engine
	detectEngine, err := frameengine.NewCanvasEngine(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("error loading rules: %v", err)
	}
	// 检测框架和组件
	detect, err := detectEngine.DetectFrameworks(ctx, index, profile.Expands)
	if err != nil {
		return nil, fmt.Errorf("error detecting frameworks and components: %v", err)
	}

	// 生成分析报告
	report := &camodels.CanvasReport{
		CodeProfile: *profile,
		Detection:   *detect,
		Timestamp:   time.Now(),
	}
	return report, nil
}
