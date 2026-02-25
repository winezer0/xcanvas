package camodels

// DetectionInfo 框架与组件识别结果 包含已检测到的框架和组件的列表。
type DetectionInfo struct {
	Frameworks []DetectedItem `json:"frameworks"`
	Components []DetectedItem `json:"components"`
}

// DetectedItem  框架与组件识别结果代表了一项已检测到的技术项目（框架或组件）。
type DetectedItem struct {
	Name     string `json:"name"`     // 例如: "gin", "log4j-core", "wails"
	Type     string `json:"type"`     // "framework" 或 "component"
	Language string `json:"language"` // 例如: "Go", "Java", "JavaScript"
	Version  string `json:"version"`  // 版本字符串，可能为空
	Category string `json:"category"` // "frontend" | "backend" | "desktop"
	Evidence string `json:"evidence"` // 人类可读的检测原因
}
