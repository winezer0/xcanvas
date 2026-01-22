package canvas

import (
	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xutils/logging"
)

// InitProjectCanvasInfo 初始化項目信息 并分析canvasReport
func InitProjectCanvasInfo(projectName, projectPath, canvasRulesDir string) *camodels.ProjectInfo {
	// 初始化项目画像信息
	projectInfo := camodels.NewEmptyProjectInfo(projectName, projectPath)
	// 获取 xcanvas 代码画像 使用 Analyze 获取语言、框架和组件信息
	canvasReport, err := Analyze(projectPath, canvasRulesDir)
	if err != nil {
		logging.Errorf("xcanvas detection failed: %v", err)
	} else {
		simpleCanvas := canvasReport.ToSimpleReport()
		//logging.Debugf("xcanvas detection simple report: %s", utils.ToJson(simpleCanvas))
		projectInfo.FilesCount = simpleCanvas.TotalFiles
		projectInfo.Languages = simpleCanvas.Languages
		projectInfo.Frameworks = simpleCanvas.Frameworks
		projectInfo.Components = simpleCanvas.Components
		projectInfo.BackendLanguages = simpleCanvas.BackendLanguages
		projectInfo.FrontendLanguages = simpleCanvas.FrontendLanguages
	}
	return projectInfo
}
