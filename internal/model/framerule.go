package model

// FrameRule 匹配组件/框架的信息 判断组件或框架是否存在
type FrameRule struct {
	// Paths: 必须存在的路径（文件或目录），全部都要存在
	Paths []string `yaml:"paths,omitempty"`

	// FileContents: 文件路径 -> 必须包含的关键字列表
	// 每个文件必须存在，且内容包含所有对应的关键字
	FileContents map[string][]string `yaml:"file_contents,omitempty"`
}

// VersionExtractor 表示一条完整的版本提取规则
type VersionExtractor struct {
	// FilePattern: 匹配的文件模式
	FilePattern string `yaml:"file_pattern"` // 匹配的文件模式
	// Patterns: 版本提取正则表达式列表
	Patterns []string `yaml:"patterns"` // 版本提取正则表达式列表
}

// Framework 内部规则模型（对应 YAML 规则文件）定义了如何检测框架或组件。在启动时从 YAML 规则文件中加载。
type Framework struct {
	Name     string             `yaml:"name"`
	Type     string             `yaml:"type"` // "framework" or "component"
	Language string             `yaml:"language"`
	Category string             `yaml:"category"` // 针对框架: "frontend"/"backend"; 针对组件: "frontend"/"backend"
	Rules    []FrameRule        `yaml:"rules"`    // 多条规则，OR 关系
	Versions []VersionExtractor `yaml:"version"`  // 多条版本提取表达式，OR 关系
}
