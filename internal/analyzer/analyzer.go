// Package analyzer 提供了 CodeCanvas 的代码画像分析功能。
package analyzer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/winezer0/xcanvas/internal/langengine"
	"github.com/winezer0/xcanvas/internal/logging"
	"github.com/winezer0/xcanvas/internal/model"
	"github.com/winezer0/xcanvas/internal/utils"
)

// CodeAnalyzer 实现代码画像分析功能。
type CodeAnalyzer struct{}

// NewCodeAnalyzer 创建一个新的代码分析器实例。
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{}
}

// AnalysisTask 定义一个分析任务
type AnalysisTask struct {
	Path    string
	LangDef *model.Language
}

// AnalysisResult 定义分析结果
type AnalysisResult struct {
	LangName string
	Stats    FileStats
	Err      error
}

var (
	extToLanguage  = make(map[string]*model.Language)
	fileToLanguage = make(map[string]*model.Language)
)

// init 初始化语言映射
func init() {
	// 初始化语言映射，直接使用新的语言规则
	for _, language := range langengine.LanguageRules {
		for _, ext := range language.Extensions {
			extToLanguage[strings.ToLower(ext)] = &language
		}
		for _, name := range language.Filenames {
			fileToLanguage[name] = &language
		}
	}
}

// AnalyzeCodeProfile 分析给定路径下的代码库并返回代码画像和文件索引。
func (a *CodeAnalyzer) AnalyzeCodeProfile(projectPath string) (*model.CodeProfile, *model.FileIndex, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, nil, err
	}
	// 初始化文件索引
	fileIndex := model.NewFileIndex(absPath)
	// 准备并发处理
	workers := autoWorkers()

	tasks := make(chan AnalysisTask, workers)
	results := make(chan AnalysisResult, workers)
	var wg sync.WaitGroup

	// 启动 Worker Pool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				stats, err := CountFileStats(task.Path)
				results <- AnalysisResult{
					LangName: task.LangDef.Name,
					Stats:    stats,
					Err:      err,
				}
			}
		}()
	}

	// 启动结果收集协程
	stats := make(map[string]*model.LangSummary)
	var errorFiles int
	done := make(chan struct{})
	go func() {
		for res := range results {
			if res.Err != nil {
				errorFiles++
				continue
			}
			summary, ok := stats[res.LangName]
			if !ok {
				summary = &model.LangSummary{Name: res.LangName}
				stats[res.LangName] = summary
			}
			summary.Count++
			summary.Code += res.Stats.Code
			summary.Comment += res.Stats.Comment
			summary.Blank += res.Stats.Blank
		}
		close(done)
	}()

	// 遍历目录并分发任务
	err = filepath.WalkDir(absPath, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			// 如果无法访问文件/目录，跳过
			return nil
		}
		if dirEntry.IsDir() {
			// 跳过隐藏目录，如 .git
			if strings.HasPrefix(dirEntry.Name(), ".") && dirEntry.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// 计算相对路径并添加到索引 (保持在主协程，无需锁)
		relPath, _ := filepath.Rel(absPath, path)
		// 统一使用 "/" 作为路径分隔符
		relPath = filepath.ToSlash(relPath)
		fileIndex.AddFile(relPath, dirEntry.Name(), filepath.Ext(dirEntry.Name()))

		// 识别语言
		langDef := extToLanguage[strings.ToLower(filepath.Ext(path))]
		if langDef == nil {
			langDef = fileToLanguage[dirEntry.Name()]
		}

		if langDef != nil {
			// 分发任务
			tasks <- AnalysisTask{
				Path:    path,
				LangDef: langDef,
			}
		}
		return nil
	})

	close(tasks)   // 停止发送任务
	wg.Wait()      // 等待所有 Worker 完成
	close(results) // 停止发送结果
	<-done         // 等待结果收集完成

	if err != nil {
		return nil, nil, err
	}

	codeProfile := convertToCodeProfile(absPath, stats, errorFiles)
	return codeProfile, fileIndex, nil
}

func autoWorkers() int {
	workers := runtime.NumCPU() / 4
	if workers < 1 {
		workers = 1
	}
	if workers > runtime.NumCPU() {
		workers = runtime.NumCPU()
	}
	return workers
}

// convertToCodeProfile converts statistics to CodeCanvas CodeProfile.
func convertToCodeProfile(absPath string, stats map[string]*model.LangSummary, errorFiles int) *model.CodeProfile {

	profile := &model.CodeProfile{
		Path:              absPath,
		TotalFiles:        0,
		TotalLines:        0,
		ErrorFiles:        errorFiles,
		FrontendLanguages: []string{},
		BackendLanguages:  []string{},
		LanguageInfos:     []model.LangInfo{},
	}

	// 将统计表转换为切片
	var summaries []model.LangSummary
	for _, summary := range stats {
		summaries = append(summaries, *summary)
	}

	for _, stat := range summaries {
		langInfo := model.LangInfo{
			Name:         stat.Name,
			Files:        int(stat.Count),
			CodeLines:    int(stat.Code),
			CommentLines: int(stat.Comment),
			BlankLines:   int(stat.Blank),
		}

		// Add to profile
		profile.LanguageInfos = append(profile.LanguageInfos, langInfo)
		profile.TotalFiles += langInfo.Files
		profile.TotalLines += langInfo.CodeLines + langInfo.CommentLines + langInfo.BlankLines
	}

	logging.Infof("profile ToJson: %s", utils.ToJson(profile))

	// 进行语言信息分析
	frontend, backend, desktop, other, allLang, expand := langengine.NewLangClassifier().DetectCategories(absPath, profile.LanguageInfos)
	profile.FrontendLanguages = frontend
	profile.BackendLanguages = backend
	profile.DesktopLanguages = desktop
	profile.OtherLanguages = other
	profile.Languages = allLang
	profile.Expands = expand
	return profile
}
