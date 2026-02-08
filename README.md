# xcanvas

轻量级、无依赖、可嵌入的代码画像引擎，面向 DevSecOps 场景，提供标准化的项目技术栈识别能力。

## 核心价值

✅ **快速构建项目画像**：输出语言、框架、组件三元组

✅ **规则驱动**：新增框架/组件仅需更新 YAML 配置，无需重新编译

## 程序运行逻辑
xcanvas 的运行逻辑分为以下几个主要步骤：

1. **命令行参数解析**：解析项目路径、规则目录、输出文件等参数
2. **初始化框架引擎**：加载嵌入式规则和用户自定义规则
3. **代码分析**：
    - 遍历项目目录，构建文件索引
    - 识别文件语言类型，统计各语言的文件数、代码行数等信息
    - 对语言进行分类（前端/后端/桌面/其他）
4. **框架和组件检测**：
    - 根据检测到的语言过滤规则
    - 使用文件索引加速匹配过程
    - 遍历规则，对每个框架进行检测，提取版本信息
5. **生成报告**：输出命令行报告和JSON格式结果


## 快速开始

### 安装

```bash
go install github.com/winezer0/xcanvas/cmd/xcanvas@latest
```

### 命令行参数

```bash
xcanvas [OPTIONS]
```

| 参数 | 长参数      | 描述        | 默认值 |
|------|----------|-----------|-----|
| -p | --path   | 项目路径      | -   |
| -r | --rules  | 规则目录      | ./rules |
| -o | --output | 输出JSON到文件 | -   |
| --lf | - | 日志文件路径 | - |
| --ll | - | 日志级别（debug/info/warn/error） | info |
| --lc | - | 控制台日志格式（TLCM OR off|null） | LM |
| -v | --version | 显示版本 | - |

### 使用示例

```bash
# 基本使用 使用内置规则
xcanvas -p /path/to/project

# 指定规则目录和输出文件
xcanvas -p /path/to/project -r /path/to/rules -o result.json

# 显示版本
xcanvas -v
```

## 规则说明

### 规则文件位置

- 默认规则：内置在二进制文件中
- 自定义规则：放在指定的规则目录中 文件扩展名应为 `.yml`
- 自定义规则会覆盖同名的默认规则

## 语言规则格式

语言规则用于识别文件的语言类型和分类，使用YAML格式定义：

```yaml
# 语言规则示例
- name: "Go"
  lineComments: ["//"]
  multiLine: [["/*", "*/"]]
  extensions: [".go"]
  filenames: []
  category: "backend"
  dynamic: []

- name: "JavaScript"
  lineComments: ["//"]
  multiLine: [["/*", "*/"]]
  extensions: [".js", ".jsx"]
  filenames: []
  category: "frontend"
  dynamic:
    - category: "backend"
      filePatterns: ["server.js", "app.js"]
      dependencies: ["express", "koa"]
```

**字段说明**：
- `name`：语言名称（如 "Go", "JavaScript"）
- `lineComments`：行注释标记
- `multiLine`：多行注释标记
- `extensions`：文件扩展名
- `filenames`：特定文件名
- `category`：默认分类（frontend/backend/desktop/other）
- `dynamic`：动态分类规则列表


### 框架/应用规则文件结构

规则文件使用 YAML 格式，每个文件可以包含多个框架/组件的检测规则。

```yaml
# 单个规则文件示例
- name: "React"
  type: "framework"
  language: "JavaScript"
  category: "frontend"
  rules:
    - paths:
        - "package.json"
      file_contents:
        package.json:
          - "react"
    - paths:
        - "src"
      file_contents:
        src/index.js:
          - "react"
  version:
    - file_pattern: "package.json"
      patterns:
        - "\"react\"\s*:\s*\"([^\"]+)\""

- name: "Vue"
  type: "framework"
  language: "JavaScript"
  category: "frontend"
  rules:
    - paths:
        - "package.json"
      file_contents:
        package.json:
          - "vue"
  version:
    - file_pattern: "package.json"
      patterns:
        - "\"vue\"\s*:\s*\"([^\"]+)\""
```

### 规则字段说明

- **name**: 框架/组件名称（必填）
- **type**: 类型，取值为 `framework` 或 `component`（必填）
- **language**: 语言（必填）
- **category**: 类别，取值为 `frontend` 或 `backend`（必填）
- **rules**: 检测规则列表（OR关系，至少一个）
  - **paths**: 必须存在的路径列表（AND关系，可为空）
  - **file_contents**: 文件内容匹配规则（AND关系，可为空）
    - **文件路径**: 必须包含的关键字列表（AND关系）
- **version**: 版本提取规则列表（OR关系，可为空）
  - **file_pattern**: 文件模式（必填）
  - **patterns**: 版本提取正则表达式列表（OR关系，至少一个）

### 规则匹配逻辑

- 多个规则文件之间：所有规则合并，同名规则会被覆盖
- 多个 `rule` 之间：OR 关系（任一规则匹配即成功）
- 单个 `rule` 内部：
  - `paths` 之间：AND 关系（所有路径必须存在）
  - `file_contents` 之间：AND 关系（所有文件必须存在且匹配）
  - `file_contents` 中单个文件的关键字之间：AND 关系（所有关键字必须存在）
- `paths` 或 `file_contents` 单个为空表示忽略该条件
- `paths` 和 `file_contents` 不能都为空

### 版本提取逻辑

- 多个版本提取规则之间：OR 关系（任一规则匹配即成功）
- 单个版本提取规则的多个正则表达式之间：OR 关系（任一正则匹配即成功）
- 正则表达式应使用捕获组提取版本号

## 技术特点

1. **高性能**：
    - 使用文件索引加速匹配过程
    - 并发处理文件分析任务
    - 文件内容缓存，避免重复读取

2. **可扩展性**：
    - 规则驱动设计，新增框架/组件仅需更新 YAML 配置
    - 支持用户自定义规则，覆盖默认规则
    - 嵌入式规则与用户规则合并机制

3. **准确性**：
    - 基于文件扩展名和文件名的语言识别
    - 动态语言分类规则，基于项目依赖和文件模式调整分类
    - 多维度规则匹配，提高框架检测准确性

4. **易用性**：
    - 简单的命令行接口
    - 清晰的输出格式
    - 支持 JSON 格式输出，便于集成到其他系统

## 贡献指南

欢迎提交Issue和Pull Request！

## 免责声明

本工具仅用于合法的安全测试和代码审计目的。使用者应当遵守相关法律法规，不得将本工具用于非法用途。开发者不承担因误用本工具而产生的任何责任。
