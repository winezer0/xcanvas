module github.com/winezer0/xcanvas

go 1.25.5

require (
	github.com/jessevdk/go-flags v1.6.1
	go.uber.org/zap v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
)

replace github.com/tree-sitter/tree-sitter-c_sharp => github.com/tree-sitter/tree-sitter-c-sharp v0.23.0

replace github.com/tree-sitter/tree-sitter-typescript/bindings/go => github.com/tree-sitter/tree-sitter-typescript v0.23.2
