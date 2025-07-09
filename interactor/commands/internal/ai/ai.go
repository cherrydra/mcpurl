package ai

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

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

func ModelContext(_ context.Context, args types.Arguments) error {
	if args.LLM == nil {
		return llm.ErrDisabled
	}
	if len(args.Args) == 0 {
		return parser.ErrInvalidUsage
	}
	switch args.Args[0] {
	case "clear":
		args.LLM.ClearContext()
		return nil
	}
	return parser.ErrInvalidUsage
}

func CallTool(ctx context.Context, args types.Arguments) error {
	if len(args.Args) == 0 {
		return parser.ErrInvalidUsage
	}
	// tool <tool> @data.json
	if len(args.Args) >= 2 && strings.HasPrefix(args.Args[1], "@") {
		data, err := parser.Parser{}.ParseData(args.Args[1])
		if err != nil {
			return fmt.Errorf("parse tool arguments: %w", err)
		}
		return args.Features.CallTool(ctx, args.Args[0], data)
	}

	// tool <tool> [options]
	flags := flag.NewFlagSet(args.Args[0], flag.ContinueOnError)
	arguments := map[string]*string{}
	tools, err := args.Features.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}
	for _, tool := range tools {
		if tool.Name != args.Args[0] {
			continue
		}
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", tool.Name)
			fmt.Fprintf(os.Stderr, "%s\n\n", tool.Description)
			fmt.Fprintln(os.Stderr, "Options:")
			flags.PrintDefaults()
		}
		for prop, v := range tool.InputSchema.Properties {
			p := new(string)
			arguments[prop] = p
			if slices.Contains(tool.InputSchema.Required, prop) {
				v.Description = fmt.Sprintf("%s (required)", cmp.Or(v.Description, v.Title))
			} else {
				v.Description = fmt.Sprintf("%s (optional)", cmp.Or(v.Description, v.Title))
			}
			flags.StringVar(p, prop, "", v.Description)
		}
	}
	if err := flags.Parse(args.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}
	params := map[string]any{}
	for k, v := range arguments {
		if *v != "" {
			params[k] = v
		}
	}
	return args.Features.CallTool1(ctx, args.Args[0], params)
}

func GetPrompt(ctx context.Context, args types.Arguments) error {
	if len(args.Args) == 0 {
		return parser.ErrInvalidUsage
	}

	// prompt <prompt> @data.json
	if len(args.Args) >= 2 && strings.HasPrefix(args.Args[1], "@") {
		data, err := parser.Parser{}.ParseData(args.Args[1])
		if err != nil {
			return fmt.Errorf("parse prompt arguments: %w", err)
		}
		return args.Features.GetPrompt(ctx, args.Args[0], data)
	}

	// prompt <prompt> [options]
	flags := flag.NewFlagSet(args.Args[0], flag.ContinueOnError)
	arguments := map[string]*string{}

	prompts, err := args.Features.ListPrompts(ctx)
	if err != nil {
		return fmt.Errorf("list prompts: %w", err)
	}
	for _, prompt := range prompts {
		if prompt.Name != args.Args[0] {
			continue
		}
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", prompt.Name)
			fmt.Fprintf(os.Stderr, "%s\n\n", prompt.Description)
			fmt.Fprintln(os.Stderr, "Options:")
			flags.PrintDefaults()
		}
		for _, prop := range prompt.Arguments {
			p := new(string)
			arguments[prop.Name] = p
			if prop.Required {
				prop.Description = fmt.Sprintf("%s (required)", prop.Description)
			} else {
				prop.Description = fmt.Sprintf("%s (optional)", prop.Description)
			}
			flags.StringVar(p, prop.Name, "", prop.Description)
		}
	}
	if err := flags.Parse(args.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}
	params := map[string]string{}
	for k, v := range arguments {
		if v != nil {
			params[k] = *v
		}
	}
	return args.Features.GetPrompt1(ctx, args.Args[0], params)
}

func ReadResource(ctx context.Context, args types.Arguments) error {
	if len(args.Args) == 0 {
		return parser.ErrInvalidUsage
	}
	return args.Features.ReadResource(ctx, args.Args[0])
}

func ListPrompts(ctx context.Context, args types.Arguments) error {
	return args.Features.PrintPrompts(ctx)
}

func ListResources(ctx context.Context, args types.Arguments) error {
	return args.Features.PrintResources(ctx)
}

func ListTools(ctx context.Context, args types.Arguments) error {
	return args.Features.PrintTools(ctx)
}
