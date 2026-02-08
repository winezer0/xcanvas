package canvas

import (
	"fmt"
	"github.com/winezer0/xutils/logging"
	"time"

	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xcanvas/internal/analyzer"
	"github.com/winezer0/xcanvas/internal/frameengine"
)

// Analyze performs a full analysis and returns a CanvasReport.
func Analyze(path string, rulesDir string) (*camodels.CanvasReport, error) {
	// 初始化框架识别规则引擎
	canvasEngine, initErr := frameengine.NewCanvasEngine(rulesDir)
	if initErr != nil {
		return nil, fmt.Errorf("init canvas engine rules error: %v", initErr)
	}

	// 分析代码结构 并重复使用 fileIndex
	codeAnalyzer := analyzer.NewCodeAnalyzer()
	codeProfile, fileIndex, analyzerErr := codeAnalyzer.AnalyzeCodeProfile(path)
	if analyzerErr != nil {
		return nil, fmt.Errorf("error analyzing code profile: %v", analyzerErr)
	}

	// 基于规则检测框架和组件 并重复使用 fileIndex
	detectInfo, detectErr := canvasEngine.DetectFrameworks(fileIndex, codeProfile.Expands)
	if detectErr != nil {
		return nil, fmt.Errorf("error detecting frameworks and components: %v", detectErr)
	}

	// 生成分析报告
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
