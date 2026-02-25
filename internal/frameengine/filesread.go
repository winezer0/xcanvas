package frameengine

import (
	"io"
	"os"
)

// GetFileContentWithCache 读取文件内容，带缓存和大文件截断（最大 5MB，只读前 1MB）
// cache 是外部传入的 map[string][]byte，用于跨调用共享缓存
func GetFileContentWithCache(path string, cache map[string][]byte) ([]byte, error) {
	if content, ok := cache[path]; ok {
		return content, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err == nil && stat.Size() > 5*1024*1024 { // >5MB
		content, err := io.ReadAll(io.LimitReader(f, 1*1024*1024)) // 读前1MB
		if err != nil {
			return nil, err
		}
		cache[path] = content
		return content, nil
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	cache[path] = content
	return content, nil
}
