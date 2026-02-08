// Package main provides the CLI interface for CodeCanvas.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/xcanvas/canvas"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/utils"
)

// Options defines the command-line parameters for CodeCanvas.
type Options struct {
	// Analysis parameters
	ProjectPath string `short:"p" long:"path" description:"path to the codebase to analyze"`
	RulesDir    string `short:"r" long:"rules" description:"detection rules dir path" default:""`
	Output      string `short:"o" long:"output" description:"write json to file"`

	// 日志参数（中文描述）
	LogFile    string `long:"lf" description:"log file path (if empty, no file will be written)"`
	LogLevel   string `long:"ll" description:"log level (debug/info/warn/error)" default:"info"`
	LogConsole string `long:"lc" description:"log format for console(TLCM OR off|null）" default:"LM"`
	Version    bool   `short:"v" long:"version" description:"show version"`
}

const (
	AppName      = "xcanvas"
	AppShortDesc = "Code fingerprint analysis"
	AppLongDesc  = "Code fingerprint analysis"
	AppVersion   = "0.1.9"
	BuildDate    = "2026-02-03"
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
	canvas.PrintCanvasReport(report)
	// 输出Json结果
	utils.SaveJSON(opts.Output, report)
}

// InitOptionsArgs 常用的工具函数，解析parser和logging配置
func InitOptionsArgs(minimumParams int) (*Options, *flags.Parser) {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = AppName
	parser.Usage = "[OPTIONS]"
	parser.ShortDescription = AppShortDesc
	parser.LongDescription = AppLongDesc

	// 命令行参数数量检查 指不包含程序名本身的参数数量
	if minimumParams > 0 && len(os.Args)-1 < minimumParams {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// 命令行参数解析检查
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Printf("Error:%v\n", err)
		os.Exit(1)
	}

	// 新增：判断是否需要显示版本信息
	if opts.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 初始化日志器
	logCfg := logging.NewLogConfig(opts.LogLevel, opts.LogFile, opts.LogConsole)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	// 处理项目路径
	if opts.ProjectPath == "" {
		logging.Fatalf("must input project path !!!")
	}

	if exists, _, _ := utils.PathExists(opts.ProjectPath); !exists {
		logging.Fatalf("project path not exists: %s !!!", opts.ProjectPath)
	}

	logging.Infof("ProjectPath: %s", opts.ProjectPath)
	return opts, parser
}
