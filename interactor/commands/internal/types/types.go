package types

import (
	"os"

	"github.com/cherrydra/mcpurl/llm"
	"github.com/cherrydra/mcpurl/mcp/features"
)

type Arguments struct {
	LLM      *llm.LLM
	Features features.ServerFeatures
	In, Out  *os.File
	Args     []string
}
