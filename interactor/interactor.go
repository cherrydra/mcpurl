package interactor

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cherrydra/mcpurl/features"
	"github.com/cherrydra/mcpurl/parser"
	"github.com/cherrydra/mcpurl/version"
	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ErrInvalidPipe = errors.New("invalid pipe command")
)

type Interactor struct {
	Session *mcp.ClientSession
}

func (i Interactor) Run(ctx context.Context) error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[36mmcpurl>\033[0m ",
		AutoComplete:    &mcpurlCompleter{ctx: ctx, s: &features.ServerFeatures{Session: i.Session}},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistoryFile:         historyFile(),
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return fmt.Errorf("create readline: %w", err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		if err := i.executeCommand(ctx, strings.TrimSpace(line)); err != nil {
			if errors.Is(err, parser.ErrInvalidUsage) {
				printUsage()
				continue
			}
			fmt.Println(err.Error())
		}
	}
	return nil
}

func (ia Interactor) executeCommand(ctx context.Context, command string) (err error) {
	// io redirect
	redirAppendParts := strings.Split(command, ">>")
	redirCreateParts := strings.Split(redirAppendParts[0], ">")
	var redirPart, redirMode string
	if len(redirAppendParts) > 1 {
		redirPart = strings.TrimSpace(redirAppendParts[len(redirAppendParts)-1])
		redirMode = "append"
	} else if len(redirCreateParts) > 1 {
		redirPart = strings.TrimSpace(redirCreateParts[len(redirCreateParts)-1])
		redirMode = "create"
	}
	stdout := os.Stdout
	switch redirMode {
	case "append":
		stdout, err = os.OpenFile(redirPart, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open file for append: %w", err)
		}
	case "create":
		stdout, err = os.Create(redirPart)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
	}

	// pipeline
	pipeParts := strings.Split(redirCreateParts[0], "|")

	var nextIn, out *os.File
	if len(pipeParts) > 1 {
		nextIn, out, err = os.Pipe()
		if err != nil {
			return fmt.Errorf("create pipe: %w", err)
		}
	}

	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ia.executeMain(ctx, strings.TrimSpace(pipeParts[0]), cmp.Or(out, stdout)); err != nil {
			errChan <- err
		}
	}()
	for i, part := range pipeParts[1:] {
		thisIn := nextIn
		thisOut := stdout
		if i < len(pipeParts)-2 {
			nextIn, thisOut, err = os.Pipe()
			if err != nil {
				errChan <- fmt.Errorf("create pipe for part %d: %w", i+1, err)
				return
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ia.executePipe(ctx, strings.TrimSpace(part), thisIn, thisOut); err != nil {
				errChan <- fmt.Errorf("execute pipe %d: %w", i+1, err)
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
		close(errChan)
	}()
	select {
	case err = <-errChan:
		return err
	case <-done:
		return nil
	}
}

func (i Interactor) executeMain(ctx context.Context, command string, out *os.File) error {
	args, err := shlex.Split(command)
	if err != nil {
		return fmt.Errorf("split command: %w", err)
	}
	if len(args) == 0 {
		return parser.ErrInvalidUsage
	}
	f := features.ServerFeatures{Session: i.Session}
	if out.Fd() != os.Stdout.Fd() {
		defer out.Close()
		f.Out = out
	}
	switch args[0] {
	case "tools":
		return f.PrintTools(ctx)
	case "prompts":
		return f.PrintPrompts(ctx)
	case "resources":
		return f.PrintResources(ctx)
	case "tool", "prompt", "resource":
		if len(args) < 2 {
			return parser.ErrInvalidUsage
		}
		var data string
		if len(args) > 2 {
			data = args[2]
		}
		after, err := parser.Parser{}.ParseData(data)
		if err != nil {
			return fmt.Errorf("parse data: %w", err)
		}
		if args[0] == "tool" {
			return f.CallTool(ctx, args[1], after)
		}
		if args[0] == "prompt" {
			return f.GetPrompt(ctx, args[1], after)
		}
		return f.ReadResource(ctx, args[1])
	case "cat":
		return i.readFile(out, args)
	case "cd":
		return i.chdir(args)
	case "clear", "cls":
		fmt.Print("\033[H\033[2J")
		return nil
	case "exit", "quit":
		os.Exit(0)
		return nil
	case "h", "help":
		printUsage()
		return nil
	case "ls":
		return i.listDir(out, args)
	case "pwd":
		return i.printPwd(out)
	case "v", "version":
		fmt.Fprintln(out, version.Short())
		return nil
	default:
		return parser.ErrInvalidUsage
	}
}

func (i Interactor) executePipe(ctx context.Context, pipePart string, in *os.File, out *os.File) error {
	defer in.Close()
	if out.Fd() != os.Stdout.Fd() {
		defer out.Close()
	}
	pipeArgs, err := shlex.Split(pipePart)
	if err != nil {
		return fmt.Errorf("split pipe command: %w", err)
	}
	if len(pipeArgs) == 0 {
		return ErrInvalidPipe
	}
	command := exec.CommandContext(ctx, pipeArgs[0], pipeArgs[1:]...)
	command.Stdin = in
	command.Stdout = out
	command.Stderr = os.Stderr
	return command.Run()
}

func (i Interactor) chdir(args []string) error {
	dir := "."
	if len(args) > 1 {
		dir = args[1]
	}
	return os.Chdir(dir)
}

func (i Interactor) listDir(out *os.File, args []string) error {
	dir := "."
	for _, arg := range args[1:] {
		if !strings.HasPrefix(arg, "-") {
			dir = arg
			break
		}
	}
	items, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}
	for _, item := range items {
		if item.IsDir() {
			fmt.Fprintf(out, "%s/\n", item.Name())
			continue
		}
		fmt.Fprintf(out, "%s\n", item.Name())
	}
	return nil
}

func (i Interactor) printPwd(out *os.File) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	fmt.Fprintln(out, dir)
	return nil
}

func (i *Interactor) readFile(out *os.File, args []string) error {
	if len(args) < 2 {
		return parser.ErrInvalidUsage
	}
	file, err := os.Open(args[1])
	if err != nil {
		return fmt.Errorf("open file %s: %w", args[1], err)
	}
	defer file.Close()
	if _, err := io.Copy(out, file); err != nil {
		return fmt.Errorf("read file %s: %w", args[1], err)
	}
	return nil
}

func printUsage() {
	fmt.Println(`Usage:
  tools                   List tools
  prompts                 List prompts
  resources               List resources
  tool <name> [data]      Call tool
  prompt <name> [data]    Get prompt
  resource <name>         Read resource

  cat <file>              Read file
  cd [dir]                Change working directory
  clear                   Clear the screen
  exit       	          Exit the interactor
  help                    Show this help message
  ls [dir]                List files in directory
  pwd                     Print current working directory
  version                 Show version information

supports command pipelining and redirection:
  tools | jq .name > tools.txt`)
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func historyFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".mcpurl_history")
}
