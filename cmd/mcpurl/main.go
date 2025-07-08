package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/cherrydra/mcpurl/features"
	"github.com/cherrydra/mcpurl/interactor"
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
	clientTransport, err := transport.Transport(parsed)
	if err != nil && !errors.Is(err, transport.ErrNoTransport) {
		return fmt.Errorf("transport: %w", err)
	}
	ctx := context.Background()
	var session *mcp.ClientSession
	if err == nil {
		client := mcp.NewClient("mcpcurl", version.Short(), nil)
		if session, err = client.Connect(ctx, clientTransport); err != nil {
			return fmt.Errorf("connect mcp server: %w", err)
		}
		defer session.Close()
	}

	var L *llm.LLM
	if parsed.LLMBaseURL != "" {
		client := openai.NewClient(
			option.WithBaseURL(parsed.LLMBaseURL),
			option.WithAPIKey(parsed.LLMApiKey),
		)
		L = &llm.LLM{
			Client: &client,
			Model:  parsed.LLMName,
		}
	}

	if parsed.Interactive {
		return (&interactor.Interactor{
			Session: session,
			Server:  strings.Join(parsed.TransportArgs(), " "),
			LLM:     L,
		}).Run(ctx)
	}

	if session == nil {
		return parser.ErrInvalidUsage
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
	if parsed.Msg != "" {
		if L == nil {
			return llm.ErrDisabled
		}
		return L.Msg(ctx, f, parsed.Msg, os.Stdout)
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
