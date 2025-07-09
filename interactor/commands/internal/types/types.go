package types

import (
	"os"

	"github.com/cherrydra/mcpurl/features"
	"github.com/cherrydra/mcpurl/llm"
)

type Arguments struct {
	LLM      *llm.LLM
	Features features.ServerFeatures
	In, Out  *os.File
	Args     []string
}
