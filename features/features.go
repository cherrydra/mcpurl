package features

import (
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerFeatures struct {
	Session *mcp.ClientSession
	Out     *os.File
}
