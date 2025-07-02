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
	transportArgs := parsed.TransportArgs()
	transportURL, err := url.Parse(transportArgs[0])
	if err != nil {
		return fmt.Errorf("parse transport url: %w", err)
	}

	var transport mcp.Transport
	switch transportURL.Scheme {
	case "", "stdio":
		cmd := cmp.Or(transportURL.Host, transportURL.Path)
		command := exec.Command(cmd, transportArgs[1:]...)
		command.Stderr = os.Stderr
		transport = mcp.NewCommandTransport(command)
	default:
		return fmt.Errorf("unsupportd transport url scheme: %s", transportURL.Scheme)
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
  stdio (standard input/output)

Accepted <options>:
  -T, --tools             list tools
  -P, --prompts           list prompts
  -R, --resources         list resources
  -t, --tool <string>     call tool
  -p, --prompt <string>   get prompt
  -r, --resource <string> read resource
  -d, --data <string>     send json data to server

  -h, --help              show this usage
  -v, --version           show version

Accepted <mcp_server> formats:
  stdio:///path/to/mcpserver [args]
  /path/to/mcpserver [args]`)
}
