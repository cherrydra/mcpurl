package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nasuci/mcpurl/features"
	"github.com/nasuci/mcpurl/interactor"
	"github.com/nasuci/mcpurl/parser"
	"github.com/nasuci/mcpurl/transport"
	"github.com/nasuci/mcpurl/version"
)

func main() {
	parser := parser.Parser{}

	runE(func() error {
		return parser.Parse(os.Args[1:])
	})

	slog.SetLogLoggerLevel(parser.LogLevel)
	if parser.Silent {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
	slog.Debug("Running in debug mode", "version", version.Short(), "go_version", version.GoVersion)

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
		json.NewEncoder(os.Stdout).Encode(map[string]string{"error": err.Error()})
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

	if parsed.Interactive {
		return interactor.Interactor{
			Session: session,
		}.Run(ctx)
	}

	f := features.ServerFeatures{Session: session}

	if parsed.Tools() {
		return f.PrintTools(ctx)
	}
	if parsed.Prompts() {
		return f.PrintPrompts(ctx)
	}
	if parsed.Resources() {
		return f.PrintResources(ctx)
	}
	if parsed.Tool() != "" {
		return f.CallTool(ctx, parsed.Tool(), parsed.Data)
	}
	if parsed.Prompt() != "" {
		return f.GetPrompt(ctx, parsed.Prompt(), parsed.Data)
	}
	if parsed.Resource() != "" {
		return f.ReadResource(ctx, parsed.Resource())
	}
	return parser.ErrInvalidUsage
}

func printUsage() {
	fmt.Println(`Usage:
  mcpurl <options> <mcp_server>

Accepted <options>:
  -T, --tools                 List tools
  -P, --prompts               List prompts
  -R, --resources             List resources
  -t, --tool <string>         Call tool
  -p, --prompt <string>       Get prompt
  -r, --resource <string>     Read resource
  -d, --data <string/@file>   Send json data to server
  -H, --header <header/@file> Pass custom header(s) to server
  -h, --help                  Show this usage
  -I, --interactive           Start interactive mode
  -l, --log-level <level>     Set log level (debug, info, warn, error)
  -s, --silent                Silent mode
  -v, --version               Show version

Accepted <mcp_server> formats:
  https://example.com/mcp [options]
  stdio:///path/to/mcpserver [args]
  /path/to/mcpserver [args]

Currently supported transports:
  http(s) (streamable http)
  stdio   (standard input/output)`)
}
