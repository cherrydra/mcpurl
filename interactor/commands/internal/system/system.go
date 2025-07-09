package system

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cherrydra/mcpurl/interactor/commands/internal/types"
	"github.com/cherrydra/mcpurl/parser"
)

func Chdir(_ context.Context, args types.Arguments) error {
	dir := "."
	if len(args.Args) > 0 {
		dir = args.Args[0]
	}
	return os.Chdir(dir)
}

func Clear(_ context.Context, args types.Arguments) error {
	fmt.Print("\033[H\033[2J")
	return nil
}

func ShowEnv(_ context.Context, args types.Arguments) error {
	for _, env := range os.Environ() {
		fmt.Fprintln(args.Out, env)
	}
	return nil
}

func ExportEnv(ctx context.Context, args types.Arguments) error {
	if len(args.Args) < 1 {
		return ShowEnv(ctx, args)
	}

	for _, arg := range args.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}
		os.Setenv(parts[0], parts[1])
	}
	return nil
}

func ListDir(_ context.Context, args types.Arguments) error {
	dir := "."
	for _, arg := range args.Args {
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
			fmt.Fprintf(args.Out, "%s/\n", item.Name())
			continue
		}
		fmt.Fprintf(args.Out, "%s\n", item.Name())
	}
	return nil
}

func PrintPwd(_ context.Context, args types.Arguments) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	fmt.Fprintln(args.Out, dir)
	return nil
}

type lastByteDetector struct {
	lastByte byte
}

func (d *lastByteDetector) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	n = len(p)
	d.lastByte = p[n-1]
	return
}

func ReadFile(_ context.Context, args types.Arguments) error {
	if len(args.Args) < 1 {
		return parser.ErrInvalidUsage
	}
	file, err := os.Open(args.Args[0])
	if err != nil {
		return fmt.Errorf("open file %s: %w", args.Args[0], err)
	}
	defer file.Close()
	detector := &lastByteDetector{}
	if _, err := io.Copy(io.MultiWriter(args.Out, detector), file); err != nil {
		return fmt.Errorf("read file %s: %w", args.Args[0], err)
	}
	if detector.lastByte != '\n' {
		fmt.Fprintln(args.Out, "\033[31m#\033[0m")
	}
	return nil
}
