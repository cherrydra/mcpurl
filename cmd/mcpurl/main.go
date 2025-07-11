package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/cherrydra/mcpurl/features"
	"github.com/cherrydra/mcpurl/interactor"
	"github.com/cherrydra/mcpurl/interactor/commands"
	"github.com/cherrydra/mcpurl/llm"
	"github.com/cherrydra/mcpurl/parser"
	"github.com/cherrydra/mcpurl/transport"
	"github.com/cherrydra/mcpurl/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	parser := parser.Parser{}

	runE(func() error {
		return parser.Parse(os.Args[1:])
	})

	slog.SetLogLoggerLevel(parser.Arguments().LogLevel)
	if parser.Arguments().Silent {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
	slog.Debug("Running in debug mode", "version", version.Short(), "go_version", version.GoVersion)

	if parser.Arguments().Help {
		printUsage()
		return
	}

	if parser.Arguments().Version {
		fmt.Println(version.GoVersion, version.Short())
		return
	}

	runE(func() error {
		return runMain(parser.Arguments())
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

func runMain(args parser.Arguments) error {
	clientTransport, err := transport.Transport(args)
	if err != nil && !errors.Is(err, transport.ErrNoTransport) {
		return fmt.Errorf("transport: %w", err)
	}
	ctx := context.Background()
	var session *mcp.ClientSession
	if err == nil {
		client := mcp.NewClient(commands.MCPImplementation, nil)
		if session, err = client.Connect(ctx, clientTransport); err != nil {
			return fmt.Errorf("connect mcp server: %w", err)
		}
		defer session.Close()
	}

	var L *llm.LLM
	if args.LLMBaseURL != "" {
		client := openai.NewClient(
			option.WithBaseURL(args.LLMBaseURL),
			option.WithAPIKey(args.LLMApiKey),
		)
		L = &llm.LLM{
			Client: &client,
			Model:  args.LLMName,
		}
	}

	commands := &commands.Commands{
		Args:    args,
		Session: session,
		LLM:     L,
	}

	if args.Interactive {
		return (&interactor.Interactor{Commands: commands}).Run(ctx)
	}

	if session == nil {
		return parser.ErrInvalidUsage
	}

	if args.Tools {
		return commands.Exec(ctx, "tools", nil, os.Stdin, os.Stdout)
	}
	if args.Prompts {
		return commands.Exec(ctx, "prompts", nil, os.Stdin, os.Stdout)
	}
	if args.Resources {
		return commands.Exec(ctx, "resources", nil, os.Stdin, os.Stdout)
	}
	if args.Tool != "" {
		return (&features.ServerFeatures{Session: commands.Session}).CallTool(ctx, args.Tool, args.Data)
	}
	if args.Prompt != "" {
		return (&features.ServerFeatures{Session: commands.Session}).GetPrompt(ctx, args.Prompt, args.Data)
	}
	if args.Resource != "" {
		return (&features.ServerFeatures{Session: commands.Session}).ReadResource(ctx, args.Resource)
	}
	if args.Msg != "" {
		return commands.Exec(ctx, "msg", []string{args.Msg}, os.Stdin, os.Stdout)
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
  -K, --llm-api-key <key>     API key for authenticating with the LLM
  -L, --llm-base-url <url>    Base URL of the LLM service
  -M, --llm-name <name>       Name of the LLM model to use
  -l, --log-level <level>     Set log level (debug, info, warn, error)
  -m, --msg <message>         Talk to LLM
  -s, --silent                Silent mode
  -v, --version               Show version

Accepted <mcp_server> formats:
  https://example.com/mcp [options]
  stdio:///path/to/mcpserver [args] (or simply /path/to/mcpserver [args])

Currently supported transports:
  http(s) (streamable http)
  stdio   (standard input/output)`)
}
