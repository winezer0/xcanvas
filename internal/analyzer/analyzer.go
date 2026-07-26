// Package analyzer 提供了 CodeCanvas 的代码画像分析功能。
package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/winezer0/xutils/progress"
	"github.com/winezer0/xutils/utils"

	"github.com/winezer0/xcanvas/camodels"
	"github.com/winezer0/xcanvas/internal/langengine"
	"github.com/winezer0/xutils/logging"
)

// contextCheckInterval controls how often ctx.Err() is polled during walk.
const contextCheckInterval = 100

// CodeAnalyzer 实现代码画像分析功能。
type CodeAnalyzer struct{}

// NewCodeAnalyzer 创建一个新的代码分析器实例。
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{}
}

// AnalysisTask 定义一个分析任务
type AnalysisTask struct {
	Path    string
	LangDef *camodels.Language
}

// AnalysisResult 定义分析结果
type AnalysisResult struct {
	LangName string
	Stats    FileStats
	Err      error
}

var (
	extToLanguage  = make(map[string]*camodels.Language)
	fileToLanguage = make(map[string]*camodels.Language)
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
// 使用默认 WalkOptions，不接受取消信号（向后兼容）。
func (a *CodeAnalyzer) AnalyzeCodeProfile(projectPath string) (*camodels.CodeProfile, *camodels.FileIndex, error) {
	profile, index, _, err := a.AnalyzeCodeProfileWithContext(
		context.Background(), projectPath, DefaultWalkOptions(),
	)
	return profile, index, err
}

// AnalyzeCodeProfileWithContext 分析给定路径下的代码库，支持取消和资源限制。
// 返回代码画像、文件索引、遍历诊断信息和可能的错误。
func (a *CodeAnalyzer) AnalyzeCodeProfileWithContext(
	ctx context.Context, projectPath string, opts WalkOptions,
) (*camodels.CodeProfile, *camodels.FileIndex, *WalkDiagnostics, error) {
	opts = opts.Normalize()

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}

	fileIndex := camodels.NewFileIndex(absPath)
	var taskList []AnalysisTask
	var diag WalkDiagnostics

	// Walk directory collecting analysis tasks with resource limits.
	err = a.walkAndCollect(ctx, absPath, opts, fileIndex, &taskList, &diag)
	if err != nil {
		return nil, nil, &diag, err
	}

	// Process collected tasks concurrently.
	stats, errorFiles := a.processTasks(taskList)

	codeProfile := convertToCodeProfile(absPath, stats, errorFiles)
	return codeProfile, fileIndex, &diag, nil
}

// walkAndCollect traverses the directory tree applying resource limits.
func (a *CodeAnalyzer) walkAndCollect(
	ctx context.Context,
	absPath string,
	opts WalkOptions,
	fileIndex *camodels.FileIndex,
	taskList *[]AnalysisTask,
	diag *WalkDiagnostics,
) error {
	fileCount := 0

	return filepath.WalkDir(absPath, func(path string, dirEntry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip inaccessible entries
		}

		// Periodic context check.
		if fileCount%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Skip symlinks unless explicitly allowed.
		if dirEntry.Type()&os.ModeSymlink != 0 {
			if !opts.FollowSymlinks {
				diag.SkippedSymlinks++
				if dirEntry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if dirEntry.IsDir() {
			// Skip hidden directories.
			if strings.HasPrefix(dirEntry.Name(), ".") && dirEntry.Name() != "." {
				return filepath.SkipDir
			}
			// Enforce depth limit.
			depth := pathDepth(absPath, path)
			if depth >= opts.MaxDepth {
				diag.MaxDepthReached = true
				return filepath.SkipDir
			}
			return nil
		}

		// Enforce file count limit.
		if fileCount >= opts.MaxFiles {
			diag.Truncated = true
			return filepath.SkipAll
		}

		// Check file size via DirEntry info.
		if info, infoErr := dirEntry.Info(); infoErr == nil {
			if info.Size() > opts.MaxFileSize {
				diag.SkippedLarge++
				return nil
			}
		}

		fileCount++

		// Compute relative path and add to index.
		relPath, _ := filepath.Rel(absPath, path)
		relPath = filepath.ToSlash(relPath)
		fileIndex.AddFile(relPath, dirEntry.Name(), filepath.Ext(dirEntry.Name()))

		// Identify language.
		langDef := extToLanguage[strings.ToLower(filepath.Ext(path))]
		if langDef == nil {
			langDef = fileToLanguage[dirEntry.Name()]
		}
		if langDef != nil {
			*taskList = append(*taskList, AnalysisTask{Path: path, LangDef: langDef})
		}
		return nil
	})
}

// processTasks runs concurrent file stats collection.
func (a *CodeAnalyzer) processTasks(taskList []AnalysisTask) (map[string]*camodels.LangSummary, int) {
	bar := progress.NewProcessBar(int64(len(taskList)), "Analyzing Code")
	workers := autoWorkers()

	tasks := make(chan AnalysisTask, len(taskList))
	results := make(chan AnalysisResult, len(taskList))
	var wg sync.WaitGroup

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
				_ = bar.Add(1)
			}
		}()
	}

	stats := make(map[string]*camodels.LangSummary)
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
				summary = &camodels.LangSummary{Name: res.LangName}
				stats[res.LangName] = summary
			}
			summary.Count++
			summary.Code += res.Stats.Code
			summary.Comment += res.Stats.Comment
			summary.Blank += res.Stats.Blank
		}
		close(done)
	}()

	for _, task := range taskList {
		tasks <- task
	}
	close(tasks)
	wg.Wait()
	close(results)
	<-done
	return stats, errorFiles
}

// pathDepth computes the relative depth of path from root (root itself = 0).
func pathDepth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(filepath.ToSlash(rel), "/") + 1
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
func convertToCodeProfile(absPath string, stats map[string]*camodels.LangSummary, errorFiles int) *camodels.CodeProfile {

	profile := &camodels.CodeProfile{
		Path:              absPath,
		TotalFiles:        0,
		TotalLines:        0,
		ErrorFiles:        errorFiles,
		FrontendLanguages: []string{},
		BackendLanguages:  []string{},
		LanguageInfos:     []camodels.LangInfo{},
	}

	// 将统计表转换为切片
	var summaries []camodels.LangSummary
	for _, summary := range stats {
		summaries = append(summaries, *summary)
	}

	for _, stat := range summaries {
		langInfo := camodels.LangInfo{
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

	logging.Infof("profile ToJson: %s", utils.ToJSON(profile))

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
