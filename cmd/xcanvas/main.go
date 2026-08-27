// Package main provides the CLI interface for CodeCanvas.
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/winezer0/xcanvas/canvas"
)

func main() {
	// 打印命令行输入配置
	opts, _ := InitOptionsArgs(1)

	// Analyze operation
	report, err := canvas.Analyze(opts.ProjectPath, opts.RulesDir, opts.logger)
	if err != nil {
		opts.logger.Error("code profile analysis failed", slog.Any("error", err))
		if closeErr := opts.managedLog.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "error: close logger: %v\n", closeErr)
		}
		os.Exit(1)
	}

	// 输出命令行报告
	PrintCanvasReport(report)
	// 输出Json结果
	saveJSON(opts.Output, report, opts.logger)
	if err := opts.managedLog.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "error: close logger: %v\n", err)
		os.Exit(1)
	}
}

// saveJSON 将结果序列化为JSON并写入文件
func saveJSON(path string, v any, loggers ...*slog.Logger) {
	if path == "" {
		return
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		if len(loggers) > 0 && loggers[0] != nil {
			loggers[0].Error("marshal JSON failed", slog.Any("error", err))
		}
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		if len(loggers) > 0 && loggers[0] != nil {
			loggers[0].Error("write JSON failed", slog.Any("error", err))
		}
	}
}
