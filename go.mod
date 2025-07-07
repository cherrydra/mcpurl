module github.com/cherrydra/mcpurl

go 1.24

toolchain go1.24.4

require (
	github.com/chzyer/readline v1.5.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/modelcontextprotocol/go-sdk v0.1.0
	github.com/openai/openai-go v1.8.2
)

require (
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace github.com/chzyer/readline v1.5.1 => github.com/cherrydra/readline v1.5.2
