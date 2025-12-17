// Package analyzer 提供了 CodeCanvas 的代码画像分析功能。
package analyzer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/winezer0/codecanvas/internal/classifier"
	"github.com/winezer0/codecanvas/internal/model"
)

// LanguageSummary 保存单一语言的统计结果
type LanguageSummary struct {
	Name    string
	Code    int64
	Comment int64
	Blank   int64
	Count   int64
}

// CodeAnalyzer 实现代码画像分析功能。
type CodeAnalyzer struct{}

// NewCodeAnalyzer 创建一个新的代码分析器实例。
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{}
}

// AnalysisTask 定义一个分析任务
type AnalysisTask struct {
	Path    string
	LangDef *LanguageDefinition
}

// AnalysisResult 定义分析结果
type AnalysisResult struct {
	LangName string
	Stats    FileStats
	Err      error
}

// AnalyzeCodeProfile 分析给定路径下的代码库并返回代码画像和文件索引。
func (a *CodeAnalyzer) AnalyzeCodeProfile(path string) (*model.CodeProfile, *model.FileIndex, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}

	// 初始化文件索引
	fileIndex := model.NewFileIndex(absPath)

	// 准备并发处理
	numWorkers := runtime.NumCPU()
	tasks := make(chan AnalysisTask, numWorkers)
	results := make(chan AnalysisResult, numWorkers)
	var wg sync.WaitGroup

	// 启动 Worker Pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				stats, err := CountFile(task.Path, task.LangDef)
				results <- AnalysisResult{
					LangName: task.LangDef.Name,
					Stats:    stats,
					Err:      err,
				}
			}
		}()
	}

	// 启动结果收集协程
	stats := make(map[string]*LanguageSummary)
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
				summary = &LanguageSummary{Name: res.LangName}
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
	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// 如果无法访问文件/目录，跳过
			return nil
		}
		if d.IsDir() {
			// 跳过隐藏目录，如 .git
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// 计算相对路径并添加到索引 (保持在主协程，无需锁)
		relPath, _ := filepath.Rel(absPath, path)
		// 统一使用 "/" 作为路径分隔符
		relPath = filepath.ToSlash(relPath)
		fileIndex.AddFile(relPath, d.Name(), filepath.Ext(d.Name()))

		// 识别语言
		ext := filepath.Ext(path)
		langDef := GetLanguageByExtension(ext)
		if langDef == nil {
			langDef = GetLanguageByFilename(d.Name())
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

	// 将统计表转换为切片
	var summaries []LanguageSummary
	for _, s := range stats {
		summaries = append(summaries, *s)
	}

	p := a.convertToCodeProfile(absPath, summaries, errorFiles)
	lc := classifier.NewLanguageClassifier()
	// 尝试加载自定义规则（如果存在）
	_ = lc.LoadFromFile(filepath.Join(absPath, "lang-rules.json"))
	frontend, backend, desktop, alls := lc.DetectCategories(absPath, p.LanguageInfos)
	p.FrontendLanguages = frontend
	p.BackendLanguages = backend
	p.DesktopLanguages = desktop
	p.Languages = alls
	p.ExpandLanguages = expandLanguages(alls)

	return p, fileIndex, nil
}

// convertToCodeProfile converts statistics to CodeCanvas CodeProfile.
func (a *CodeAnalyzer) convertToCodeProfile(path string, results []LanguageSummary, errorFiles int) *model.CodeProfile {
	profile := &model.CodeProfile{
		Path:              path,
		TotalFiles:        0,
		TotalLines:        0,
		ErrorFiles:        errorFiles,
		FrontendLanguages: []string{},
		BackendLanguages:  []string{},
		LanguageInfos:     []model.LanguageInfo{},
	}

	for _, stat := range results {
		langInfo := model.LanguageInfo{
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

	return profile
}
