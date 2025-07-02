package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nasuci/mcpurl/parser"
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

func runMain(parser parser.Parser) error {
	transportArgs := parser.TransportArgs()
	transportURL, err := url.Parse(transportArgs[0])
	if err != nil {
		return fmt.Errorf("parse transport url: %w", err)
	}

	var transport *mcp.CommandTransport
	switch transportURL.Scheme {
	case "", "stdio":
		cmd := cmp.Or(transportURL.Host, transportURL.Path)
		transport = mcp.NewCommandTransport(exec.Command(cmd, transportArgs[1:]...))
	default:
		return fmt.Errorf("unsupportd transport url scheme: %s", transportURL.Scheme)
	}

	ctx := context.Background()
	client := mcp.NewClient("mcpcurl", "v0.1", nil)
	session, err := client.Connect(ctx, transport)
	if err != nil {
		return fmt.Errorf("connect mcp server: %w", err)
	}
	defer session.Close()
	if parser.Tools() {
		return listTools(ctx, session)
	}
	if parser.Prompts() {
		return listPrompts(ctx, session)
	}
	if parser.Resources() {
		return listResources(ctx, session)
	}
	if parser.Tool() != "" {
		return callTool(ctx, session, parser.Tool(), parser.Data())
	}
	if parser.Prompt() != "" {
		return getPrompt(ctx, session, parser.Prompt(), parser.Data())
	}
	if parser.Resource() != "" {
		return readResource(ctx, session, parser.Resource())
	}
	return errors.New("invalid usage")
}

func printUsage() {
	fmt.Println(`Usage:
  mcpurl <options> <mcp_server>

Accepted <options>:
  -t, --tools             list tools
  -p, --prompts           list prompts
  -r, --resources         list resources
  -T, --tool <string>     call tool
  -P, --prompt <string>   get prompt
  -R, --resource <string> read resource

  -h, --help              show this usage

Currently supported transport:
  stdio (standard input/output)

Accepted <mcp_server> formats:
  stdio:///path/to/mcpserver [args]   # Explicit stdio scheme
  /path/to/mcpserver [args]           # Implicit stdio scheme`)
}
