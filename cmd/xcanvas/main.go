// Package main provides the CLI interface for CodeCanvas.
package main

import (
	"encoding/json"
	"os"

	"github.com/winezer0/slogs"

	"github.com/winezer0/xcanvas/canvas"
)

func main() {
	// 打印命令行输入配置
	opts, _ := InitOptionsArgs(1)

	// Analyze operation
	report, err := canvas.Analyze(opts.ProjectPath, opts.RulesDir)
	if err != nil {
		slogs.Errorf("Error analyzing code profile: %v\n", err)
		os.Exit(1)
	}

	// 输出命令行报告
	PrintCanvasReport(report)
	// 输出Json结果
	saveJSON(opts.Output, report)
}

// saveJSON 将结果序列化为JSON并写入文件
func saveJSON(path string, v any) {
	if path == "" {
		return
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		slogs.Errorf("marshal json error: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		slogs.Errorf("write json file error: %v", err)
	}
}
