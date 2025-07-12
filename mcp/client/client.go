package client

import (
	"github.com/cherrydra/mcpurl/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	Implementation = &mcp.Implementation{
		Name:    "mcpurl",
		Title:   "Command Line General AI Agent",
		Version: version.Short(),
	}
)
