package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cherrydra/mcpurl/features"
	"github.com/cherrydra/mcpurl/interactor/commands/internal/ai"
	"github.com/cherrydra/mcpurl/interactor/commands/internal/system"
	"github.com/cherrydra/mcpurl/interactor/commands/internal/types"
	"github.com/cherrydra/mcpurl/llm"
	"github.com/cherrydra/mcpurl/parser"
	"github.com/cherrydra/mcpurl/transport"
	"github.com/cherrydra/mcpurl/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	MCPImplementation = &mcp.Implementation{
		Name:    "mcpurl",
		Title:   "Command Line General AI Agent",
		Version: version.Short(),
	}

	registry = map[string]func(ctx context.Context, args types.Arguments) error{}
)

func init() {
	registry["ctx"] = ai.ModelContext
	registry["msg"] = ai.Msg
	registry["p"] = ai.GetPrompt
	registry["prompt"] = ai.GetPrompt
	registry["P"] = ai.ListPrompts
	registry["prompts"] = ai.ListPrompts
	registry["r"] = ai.ReadResource
	registry["resource"] = ai.ReadResource
	registry["R"] = ai.ListResources
	registry["resources"] = ai.ListResources
	registry["t"] = ai.CallTool
	registry["tool"] = ai.CallTool
	registry["T"] = ai.ListTools
	registry["tools"] = ai.ListTools

	registry["cat"] = system.ReadFile
	registry["cd"] = system.Chdir
	registry["clear"] = system.Clear
	registry["env"] = system.ShowEnv
	registry["export"] = system.ExportEnv
	registry["ls"] = system.ListDir
	registry["pwd"] = system.PrintPwd
}

type Commands struct {
	Args    parser.Arguments
	Session *mcp.ClientSession
	LLM     *llm.LLM

	connectedServer string
}

func (c *Commands) Exec(ctx context.Context, command string, args []string, in, out *os.File) error {
	switch command {
	case "c", "connect":
		return c.connect(ctx, args, out)
	case "disconnect":
		return c.disconnect(ctx, out)
	case "s", "status":
		return c.showStatus(ctx, out)
	case "q", "exit":
		return os.ErrProcessDone
	case "h", "help":
		return c.PrintUsage()
	case "v", "version":
		fmt.Fprintln(out, version.Short())
		return nil
	}

	if cmd, ok := registry[command]; ok {
		return cmd(ctx, types.Arguments{
			LLM:      c.LLM,
			Features: features.ServerFeatures{Session: c.Session, Out: out},
			In:       in,
			Out:      out,
			Args:     args,
		})
	}

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Commands) Close() error {
	if c.Session != nil {
		c.Session.Close()
		c.Session = nil
	}
	return nil
}

func (i *Commands) connect(ctx context.Context, args []string, out *os.File) error {
	if len(args) == 0 {
		return parser.ErrInvalidUsage
	}

	parsed := parser.Parser{}
	if err := parsed.Parse(args); err != nil {
		return fmt.Errorf("parse transport args: %w", err)
	}
	parsedArgs := parsed.Arguments()
	parsedArgs.Silent = true
	clientTransport, err := transport.Transport(parsedArgs)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}
	client := mcp.NewClient(MCPImplementation, nil)
	session, err := client.Connect(ctx, clientTransport)
	if err != nil {
		return fmt.Errorf("connect mcp server: %w", err)
	}
	if i.Session != nil {
		json.NewEncoder(out).Encode(map[string]string{"msg": "disconnecting"})
		i.Session.Close()
	}
	i.Session = session
	i.connectedServer = strings.Join(parsedArgs.TransportArgs, " ")
	return i.showStatus(ctx, out)
}

func (i *Commands) disconnect(ctx context.Context, out *os.File) error {
	if i.Session == nil {
		return nil
	}
	json.NewEncoder(out).Encode(map[string]string{"msg": "disconnecting"})
	i.Session.Close()
	i.Session = nil
	i.connectedServer = ""
	return i.showStatus(ctx, out)
}

func (i *Commands) showStatus(ctx context.Context, out *os.File) error {
	status := features.ErrNoSession.Error()
	if i.Session != nil {
		if sid := i.Session.ID(); sid != "" {
			status = fmt.Sprintf("connected (%s)", sid)
		} else {
			status = "connected"
		}
		if err := i.Session.Ping(ctx, nil); err != nil {
			status = "unhealth"
		}
		if i.connectedServer == "" {
			i.connectedServer = strings.Join(i.Args.TransportArgs, " ")
		}
	}
	json.NewEncoder(out).Encode(struct {
		LLM    string `json:"llm,omitzero"`
		Model  string `json:"model,omitzero"`
		Server string `json:"server,omitzero"`
		Status string `json:"status,omitzero"`
	}{i.Args.LLMBaseURL, i.Args.LLMName, i.connectedServer, status})
	return nil
}

func (c *Commands) PrintUsage() error {
	fmt.Println(`Available Commands:
  tools                           List tools
  prompts                         List prompts
  resources                       List resources
  tool <name> [options]           Call tool
  prompt <name> [options]         Get prompt
  resource <name>                 Read resource
  ctx <subcmd>                    LLM context operations
  msg <message>                   Talk to LLM
  connect <mcp_server> [options]  Connect to server
  disconnect                      Disconnect from server
  status                          Show connection info

System Commands:
  cat <file>                      Read file
  cd [dir]                        Change working directory
  clear                           Clear the screen
  export [name=value ...]         Set/get environment variables
  exit                            Exit the interactor
  help                            Show this help message
  ls [dir]                        List files in directory
  pwd                             Print working directory
  version                         Show version information

Supports command pipelining and stdout redirection:
  tools | jq .name > tools.txt`)
	return nil
}
