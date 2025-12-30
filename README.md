# CodeCanvas

轻量级、无依赖、可嵌入的代码画像引擎，面向 DevSecOps 场景，提供标准化的项目技术栈识别能力。

## 核心价值

✅ **快速构建项目画像**：输出语言、框架、组件三元组

✅ **规则驱动**：新增框架/组件仅需更新 YAML 配置，无需重新编译


## 快速开始

### 安装

```bash
go install github.com/winezer0/xcanvas/cmd/codecanvas@latest
```

## 规则说明
rules规则说明： 
- 多个 rule之间是OR关系 
- Rule内部是AND关系 (paths和file_contents, paths之间, file_contents之间, file_contents的文件关键字之间)
- paths或file_contents单个为空表示忽略 
- paths和file_contents不能都为空

```
rules:
  # 通过go.mod文件内容检测
  - paths:
    file_contents:
      go.mod:
        - "github.com/wailsapp/wails/v2"
  # 通过wails.json或wails.toml文件检测
  - paths:
      - "wails.json"
  - paths:
      - "wails.toml"
  # 通过main.go或app.go文件检测
  - paths:
    file_contents:
      main.go:
        - "\"github.com/wailsapp/wails/v2\""
  - paths:
    file_contents:
      app.go:
        - "\"github.com/wailsapp/wails/v2\""
        
```
version规则说明：
- 多个规则之间是OR关系 
- 多个patterns之间是OR关系 
- 目标是匹配到一个版本号即可 为空时不进行匹配
```
version:
  - file_pattern: "go.mod"
    patterns:
      - "github.com/wailsapp/wails/v2\s+v([\d.]+)"
```