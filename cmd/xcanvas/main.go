// Package main provides the CLI interface for CodeCanvas.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/xcanvas/canvas"
	"github.com/winezer0/xcanvas/internal/logging"
	"github.com/winezer0/xcanvas/internal/utils"
)

// Options defines the command-line parameters for CodeCanvas.
type Options struct {
	// Analysis parameters
	Path     string `short:"p" long:"path" description:"Path to the codebase to analyze"`
	RulesDir string `short:"r" long:"rules" description:"Directory containing detection RulesDirDir" default:"./rules"`
	Output   string `short:"o" long:"output" description:"Write JSON to path or URL"`

	// 日志参数（中文描述）
	LogFile       string `long:"lf" description:"Log file path (if empty, no file will be written)"`
	LogLevel      string `long:"ll" description:"log level (debug/info/warn/error)" default:"info"`
	ConsoleFormat string `long:"cf" description:"console log format (TLCM OR off|null）" default:"CM"`
	Version       bool   `short:"v" long:"version" description:"show version"`
}

const (
	AppName      = "codecanvas"
	AppShortDesc = "Code fingerprint analysis"
	AppLongDesc  = "Code fingerprint analysis"
	AppVersion   = "0.0.9"
	BuildDate    = "2025-12-30"
)

func main() {
	// 解析命令行参数
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS]"
	parser.ShortDescription = AppShortDesc
	parser.LongDescription = AppLongDesc

	// 命令行參數解析
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			return
		}
		fmt.Printf("options parsed error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logCfg := logging.NewLogConfig(opts.LogLevel, opts.LogFile, opts.ConsoleFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("Init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	// 新增：判断是否需要显示版本信息
	if opts.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 进行路径分析
	if opts.Path != "" {
		// Analyze operation
		report, err := canvas.Analyze(opts.Path, opts.RulesDir)
		if err != nil {
			fmt.Printf("Error analyzing code profile: %v\n", err)
			os.Exit(1)
		}

		// 输出Json结果
		if opts.Output != "" {
			utils.EnsureDir(opts.Output, true)
			if err := utils.WriteJSON(opts.Output, report); err != nil {
				fmt.Printf("Error writing output: %v\n", err)
				os.Exit(1)
			}
		}

		// 输出命令行报告
		canvas.PrintReport(report)
	}
}
