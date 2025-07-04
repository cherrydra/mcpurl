package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nasuci/mcpurl/cmd/mcpurl/transport"
	"github.com/nasuci/mcpurl/parser"
	"github.com/nasuci/mcpurl/version"
)

func main() {
	parser := parser.Parser{}
	runE(func() error {
		return parser.Parse(os.Args[1:])
	})
	if parser.Help {
		printUsage()
		return
	}
	if parser.Version {
		fmt.Println(version.GoVersion, version.Short())
		return
	}
	runE(func() error {
		return runMain(parser)
	})
}

func runE(run func() error) {
	err := run()
	if errors.Is(err, parser.ErrInvalidUsage) {
		printUsage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func runMain(parsed parser.Parser) error {
	transport, err := transport.Transport(parsed)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}
	ctx := context.Background()
	client := mcp.NewClient("mcpcurl", version.Short(), nil)
	session, err := client.Connect(ctx, transport)
	if err != nil {
		return fmt.Errorf("connect mcp server: %w", err)
	}
	defer session.Close()
	if parsed.Tools() {
		return listTools(ctx, session)
	}
	if parsed.Prompts() {
		return listPrompts(ctx, session)
	}
	if parsed.Resources() {
		return listResources(ctx, session)
	}
	if parsed.Tool() != "" {
		return callTool(ctx, session, parsed.Tool(), parsed.Data())
	}
	if parsed.Prompt() != "" {
		return getPrompt(ctx, session, parsed.Prompt(), parsed.Data())
	}
	if parsed.Resource() != "" {
		return readResource(ctx, session, parsed.Resource())
	}
	return parser.ErrInvalidUsage
}

func printUsage() {
	fmt.Println(`Usage:
  mcpurl <options> <mcp_server>

Currently supported transport:
  http(s) (streamable http)
  stdio (standard input/output)

Accepted <options>:
  -T, --tools             List tools
  -P, --prompts           List prompts
  -R, --resources         List resources
  -t, --tool <string>     Call tool
  -p, --prompt <string>   Get prompt
  -r, --resource <string> Read resource
  -d, --data <string>     Send json data to server
  -H, --header <header>   Pass custom header(s) to server

  -h, --help              Show this usage
  -v, --version           Show version

Accepted <mcp_server> formats:
  https://example.com/mcp [options]
  stdio:///path/to/mcpserver [args]
  /path/to/mcpserver [args]`)
}
