// Package main provides the CLI interface for CodeCanvas.
package main

import (
	"github.com/winezer0/xcanvas/canvas"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/utils"
)

func main() {
	// 打印命令行输入配置
	opts, _ := InitOptionsArgs(1)

	// Analyze operation
	report, err := canvas.Analyze(opts.ProjectPath, opts.RulesDir)
	if err != nil {
		logging.Fatalf("Error analyzing code profile: %v\n", err)
	}

	// 输出命令行报告
	PrintCanvasReport(report)
	// 输出Json结果
	utils.SaveJSON(opts.Output, report)
}
