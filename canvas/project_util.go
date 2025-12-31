package canvas

import (
	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xcanvas/internal/logging"
	"github.com/winezer0/xcanvas/internal/utils"
)

// InitProjectCanvasInfo 初始化項目信息 并分析canvasReport
func InitProjectCanvasInfo(projectPath string, rulesDir string, fileCounts int) *camodels.ProjectInfo {
	// 初始化项目画像信息
	projectInfo := &camodels.ProjectInfo{
		ProjectPath: projectPath,
		FilesCount:  fileCounts,
	}

	// 获取 xcanvas 代码画像 使用 Analyze 获取语言、框架和组件信息
	canvasReport, err := Analyze(projectPath, rulesDir)
	if err != nil {
		logging.Errorf("xcanvas detection failed: %v", err)
	} else {
		simpleCanvas := canvasReport.ToSimpleReport()
		logging.Infof("xcanvas detection simple report: %s", utils.ToJson(simpleCanvas))
		projectInfo.Languages = simpleCanvas.Languages
		projectInfo.Frameworks = simpleCanvas.Frameworks
		projectInfo.Components = simpleCanvas.Components
		projectInfo.BackendLanguages = simpleCanvas.BackendLanguages
		projectInfo.FrontendLanguages = simpleCanvas.FrontendLanguages
	}
	return projectInfo
}
