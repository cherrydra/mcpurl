package interactor

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/cherrydra/mcpurl/mcp/features"
	"github.com/google/shlex"
	"github.com/mcpurl/readline"
)

var (
	_ readline.AutoCompleter = (*mcpurlCompleter)(nil)
)

type mcpurlCompleter struct {
	ctx     context.Context
	session func() *features.ServerFeatures

	once      sync.Once
	completer *readline.PrefixCompleter
}

func (c *mcpurlCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	c.once.Do(func() {
		c.completer = readline.NewPrefixCompleter(
			readline.PcItem("tools"),
			readline.PcItem("prompts"),
			readline.PcItem("resources"),
			readline.PcItem("tool", readline.PcItemDynamic(
				c.listTools,
				readline.PcItemDynamic(func(s string) []string { return searchFiles(s, "@", FILE_SEARCH_MODE_ONLY_FILES) })),
			),
			readline.PcItem("prompt", readline.PcItemDynamic(
				c.listPrompts,
				readline.PcItemDynamic(func(s string) []string { return searchFiles(s, "@", FILE_SEARCH_MODE_ONLY_FILES) })),
			),
			readline.PcItem("resource", readline.PcItemDynamic(c.listResources)),
			readline.PcItem("msg"),
			readline.PcItem("ctx",
				readline.PcItem("clear"),
				readline.PcItem("dump"),
				readline.PcItem("del"),
				readline.PcItem("ls"),
				readline.PcItem("new"),
				readline.PcItem("use"),
			),
			readline.PcItem("connect"),
			readline.PcItem("disconnect"),
			readline.PcItem("status"),
			readline.PcItem("cat", readline.PcItemDynamic(func(s string) []string {
				return searchFiles(s, "", FILE_SEARCH_MODE_ONLY_FILES)
			})),
			readline.PcItem("cd", readline.PcItemDynamic(func(s string) []string {
				return searchFiles(s, "", FILE_SEARCH_MODE_ONLY_DIRS)
			})),
			readline.PcItem("clear"),
			readline.PcItem("exit"),
			readline.PcItem("export"),
			readline.PcItem("env"),
			readline.PcItem("help"),
			readline.PcItem("ls", readline.PcItemDynamic(func(s string) []string {
				return searchFiles(s, "", FILE_SEARCH_MODE_ONLY_DIRS)
			})),
			readline.PcItem("pwd"),
			readline.PcItem("version"),
		)
	})
	return c.completer.Do(line, pos)
}

func (c *mcpurlCompleter) listTools(prefix string) (ret []string) {
	args, _ := shlex.Split(prefix)
	tools, err := c.session().ListTools(c.ctx)
	if err != nil {
		return nil
	}
	for _, tool := range tools {
		if len(args) > 1 && !strings.HasPrefix(tool.Name, args[1]) {
			continue
		}
		ret = append(ret, tool.Name)
	}
	return
}

func (c *mcpurlCompleter) listPrompts(prefix string) (ret []string) {
	args, _ := shlex.Split(prefix)
	prompts, err := c.session().ListPrompts(c.ctx)
	if err != nil {
		return nil
	}
	for _, prompt := range prompts {
		if len(args) > 1 && !strings.HasPrefix(prompt.Name, args[1]) {
			continue
		}
		ret = append(ret, prompt.Name)
	}
	return
}

func (c *mcpurlCompleter) listResources(prefix string) (ret []string) {
	args, _ := shlex.Split(prefix)
	resources, err := c.session().ListResources(c.ctx)
	if err != nil {
		return nil
	}
	for _, resource := range resources {
		if len(args) > 1 && !strings.HasPrefix(resource.Name, args[1]) {
			continue
		}
		ret = append(ret, resource.Name)
	}
	return
}

var (
	FILE_SEARCH_MODE_ONLY_FILES int8 = 0
	FILE_SEARCH_MODE_ONLY_DIRS  int8 = 1
	FILE_SEARCH_MODE_BOTH       int8 = 2
)

func searchFiles(s, prefix string, mode int8) (ret []string) {
	var lastArg string
	if !strings.HasSuffix(s, " ") {
		args, _ := shlex.Split(s)
		if len(args) < 2 {
			return nil
		}
		lastArg = args[len(args)-1]
	}

	if !strings.HasPrefix(lastArg, prefix) {
		return nil
	}

	files, err := os.ReadDir(".")
	if err != nil {
		return nil
	}
	for _, file := range files {
		switch mode {
		case 0: // only files
			if file.IsDir() {
				continue
			}
		case 1: // only directories
			if !file.IsDir() {
				continue
			}
		default:
		}
		ret = append(ret, prefix+file.Name())
	}
	return
}
