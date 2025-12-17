package model

// FileIndex 存储代码库的文件索引结构，用于加速查找。
type FileIndex struct {
	// RootDir 是被索引的根目录绝对路径
	RootDir string
	// Files 存储所有文件的相对路径列表
	Files []string
	// NameMap 映射文件名到 Files 切片中的索引列表 (例如: "package.json" -> [0, 5, 10])
	NameMap map[string][]int
	// ExtensionMap 映射文件扩展名到 Files 切片中的索引列表 (例如: ".go" -> [1, 2, 3])
	ExtensionMap map[string][]int
}

// NewFileIndex 创建一个新的空索引
func NewFileIndex(rootDir string) *FileIndex {
	return &FileIndex{
		RootDir:      rootDir,
		Files:        make([]string, 0),
		NameMap:      make(map[string][]int),
		ExtensionMap: make(map[string][]int),
	}
}

// AddFile 向索引中添加一个文件
func (fi *FileIndex) AddFile(relPath string, fileName string, ext string) {
	idx := len(fi.Files)
	fi.Files = append(fi.Files, relPath)

	fi.NameMap[fileName] = append(fi.NameMap[fileName], idx)
	fi.ExtensionMap[ext] = append(fi.ExtensionMap[ext], idx)
}
