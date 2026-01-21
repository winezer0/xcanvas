package camodels

// ProjectInfo 封装运行级（项目级）的自定义扩展属性，存储整个项目的画像信息 和其他项目信息
type ProjectInfo struct {
	ProjectPath       string            `json:"projectPath,omitempty"` // 项目路径（自定义扩展字段，）
	Languages         []string          `json:"languages,omitempty"`   // 项目级代码语言列表（自定义扩展字段，，）
	BackendLanguages  []string          `json:"backendLanguages,omitempty"`
	FrontendLanguages []string          `json:"frontendLanguages,omitempty"`
	Frameworks        map[string]string `json:"frameworks,omitempty"` // 项目级框架列表（自定义扩展字段，）
	Components        map[string]string `json:"components,omitempty"` // 项目级组件列表（自定义扩展字段，，）
	FilesCount        int               `json:"filesCount,omitempty"` // 记录文件数量
}

// NewEmptyProjectInfo 创建并返回一个默认初始化的 ProjectInfo 实例。
func NewEmptyProjectInfo(s string) *ProjectInfo {
	return &ProjectInfo{
		ProjectPath:       s,
		Languages:         []string{},
		BackendLanguages:  []string{},
		FrontendLanguages: []string{},
		Frameworks:        map[string]string{},
		Components:        map[string]string{},
		FilesCount:        0,
	}
}
