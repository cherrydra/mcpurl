package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/cherrydra/mcpurl/interactor/commands/internal/types"
	"github.com/cherrydra/mcpurl/llm"
	"github.com/cherrydra/mcpurl/parser"
)

func Msg(ctx context.Context, args types.Arguments) error {
	if len(args.Args) == 0 {
		return parser.ErrInvalidUsage
	}
	if args.LLM == nil {
		return llm.ErrDisabled
	}
	msg := args.Args[0]
	if msg != "-" {
		return args.LLM.Msg(ctx, args.Features, msg, args.Out)
	}
	b, err := io.ReadAll(args.In)
	if err != nil {
		return fmt.Errorf("read message from stdin: %w", err)
	}
	return args.LLM.Msg(ctx, args.Features, string(b), args.Out)
}

// ModelContext handles LLM context operations.
// ctx clear - clears the current context.
// ctx dump - dumps the current context messages.
// ctx new - creates a new context.
// ctx del <index> - deletes a context by index.
// ctx ls - lists all contexts.
// ctx use <index> - switches to a context by index.
func ModelContext(_ context.Context, args types.Arguments) error {
	if args.LLM == nil {
		return llm.ErrDisabled
	}
	if len(args.Args) == 0 {
		return parser.ErrInvalidUsage
	}

	switch args.Args[0] {
	case "pop":
		msg, err := args.LLM.ContextManger.Pop()
		if err != nil {
			return err
		}
		_ = json.NewEncoder(args.Out).Encode(msg)
	case "clear":
		args.LLM.ContextManger.Clear()
	case "dump":
		for _, msg := range args.LLM.ContextManger.Current().Messages {
			b, _ := msg.MarshalJSON()
			fmt.Fprintln(args.Out, string(b))
		}
	case "new":
		args.LLM.ContextManger.New()
	case "del", "delete":
		if len(args.Args) < 2 {
			return parser.ErrInvalidUsage
		}
		index, err := strconv.ParseInt(args.Args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("parse context index: %w", err)
		}
		return args.LLM.ContextManger.Delete(int(index))
	case "ls", "list":
		ctxs := args.LLM.ContextManger.List()
		if len(ctxs) == 0 {
			fmt.Fprintln(args.Out, "No data")
			return nil
		}
		for _, ctx := range ctxs {
			if ctx.Current {
				fmt.Fprint(args.Out, "*")
			} else {
				fmt.Fprint(args.Out, " ")
			}
			fmt.Fprintf(args.Out, " %d: %s\n", ctx.Index, ctx.Title)
		}
	case "switch", "use":
		if len(args.Args) < 2 {
			return parser.ErrInvalidUsage
		}
		index, err := strconv.ParseInt(args.Args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("parse context index: %w", err)
		}
		return args.LLM.ContextManger.SwitchTo(int(index))
	default:
		return parser.ErrInvalidUsage
	}
	return nil
}
