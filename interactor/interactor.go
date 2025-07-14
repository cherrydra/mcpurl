package interactor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/cherrydra/mcpurl/interactor/commands"
	"github.com/cherrydra/mcpurl/mcp/features"
	"github.com/cherrydra/mcpurl/parser"
	"github.com/google/shlex"
	"github.com/mcpurl/readline"
)

var (
	ErrInvalidPipe = errors.New("invalid pipe command")
)

type Interactor struct {
	Commands *commands.Commands

	completer *mcpurlCompleter
}

func (i *Interactor) Run(ctx context.Context) error {
	if i.Commands == nil {
		return fmt.Errorf("commands is nil")
	}

	// restore llm contexts if running in interactive mode
	if i.Commands.LLM != nil {
		// load llm contexts from file
		if err := i.Commands.LLM.ContextManger.LoadOnce(i.Commands.Args.LLMContextFile); err != nil {
			return fmt.Errorf("load llm contexts: %w", err)
		}

		// new interactor session new context
		if !i.Commands.LLM.ContextManger.Current().IsEmpty() {
			i.Commands.LLM.ContextManger.New()
		}

		// save llm contexts on exit
		defer func() {
			if err := i.Commands.LLM.ContextManger.Save(i.Commands.Args.LLMContextFile); err != nil {
				slog.Warn("save llm contexts", "error", err)
			}
		}()
	}

	i.completer = &mcpurlCompleter{
		ctx:     ctx,
		session: func() *features.ServerFeatures { return &features.ServerFeatures{Session: i.Commands.Session} },
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[36mmcpurl>\033[0m ",
		AutoComplete:    i.completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistoryFile:         i.Commands.Args.HistoryFile,
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return fmt.Errorf("create readline: %w", err)
	}

	defer l.Close()
	defer i.Commands.Close()

	var executionCtx context.Context
	var executionCancel context.CancelFunc

	readline.CaptureExitSignal(func() {
		if executionCancel != nil {
			executionCancel()
			return
		}
		l.Close()
	})

readLoop:
	for {
		line, err := l.Readline()
		if err == io.EOF {
			break
		}
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		}
		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		executionCtx, executionCancel = context.WithCancel(ctx)
		err = i.executeCommand(executionCtx, command)
		executionCancel()
		executionCancel = nil
		switch err {
		case parser.ErrInvalidUsage:
			_ = i.Commands.PrintUsage()
		case nil, context.Canceled: // just ignore
		case os.ErrProcessDone:
			break readLoop
		default:
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	}
	return nil
}

func (ia *Interactor) executeCommand(ctx context.Context, command string) (err error) {
	// io redirect
	stdout := os.Stdout
	stdin := os.Stdin
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
	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	for i, part := range pipeParts {
		thisIn := stdin
		thisOut := stdout
		if i < len(pipeParts)-1 {
			stdin, thisOut, err = os.Pipe()
			if err != nil {
				errChan <- fmt.Errorf("create pipe for part %d: %w", i+1, err)
				return
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if thisIn.Fd() != os.Stdin.Fd() {
					thisIn.Close()
				}
				if thisOut.Fd() != os.Stdout.Fd() {
					thisOut.Close()
				}
			}()
			args, err := shlex.Split(part)
			if err != nil {
				errChan <- fmt.Errorf(`parse "%s": %w`, strings.TrimSpace(part), err)
			}
			if err := ia.Commands.Exec(ctx, args[0], args[1:], thisIn, thisOut); err != nil {
				errChan <- err
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

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
