package parser

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidUsage = errors.New("invalid usage")
)

type Arguments struct {
	// Data
	Data          string
	Headers       []string
	LogLevel      slog.Level
	LLMBaseURL    string
	LLMApiKey     string
	LLMName       string
	Silent        bool
	TransportArgs []string

	// Actions
	Help        bool
	Interactive bool
	Msg         string
	Prompt      string
	Prompts     bool
	Resource    string
	Resources   bool
	Tool        string
	Tools       bool
	Version     bool

	HistoryFile    string
	LLMContextFile string
}

type Parser struct {
	args Arguments
}

func (p *Parser) Parse(args []string) error {
	if err := p.applyFromEnv(); err != nil {
		return fmt.Errorf("apply from env: %w", err)
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-T", "--tools":
			p.args.Tools = true
		case "-P", "--prompts":
			p.args.Prompts = true
		case "-R", "--resources":
			p.args.Resources = true
		case "-h", "--help":
			p.args.Help = true
			return nil
		case "-I", "--interactive":
			p.args.Silent = true
			p.args.Interactive = true
		case "-s", "--silent":
			p.args.Silent = true
		case "-v", "--version":
			p.args.Version = true
			return nil
		default:
			switch arg {
			case "-t", "--tool", "-p", "--prompt", "-r", "--resource", "-d", "--data", "-H", "--header", "-l", "--log-level",
				"-K", "--llm-api-key", "-L", "--llm-base-url", "-M", "--llm-name", "-m", "--msg":
				if len(args) < i+2 {
					return ErrInvalidUsage
				}
				switch arg {
				case "-t", "--tool":
					p.args.Tool = args[i+1]
				case "-p", "--prompt":
					p.args.Prompt = args[i+1]
				case "-r", "--resource":
					p.args.Resource = args[i+1]
				case "-d", "--data":
					data, err := p.ParseData(args[i+1])
					if err != nil {
						return fmt.Errorf("parse data: %w", err)
					}
					p.args.Data = data
				case "-H", "--header":
					headers, err := p.ParseHeader(args[i+1])
					if err != nil {
						return fmt.Errorf("parse header: %w", err)
					}
					p.args.Headers = append(p.args.Headers, headers...)
				case "-l", "--log-level":
					if err := p.args.LogLevel.UnmarshalText([]byte(args[i+1])); err != nil {
						return fmt.Errorf("parse log level: %w", err)
					}
				case "-K", "--llm-api-key":
					p.args.LLMApiKey = args[i+1]
				case "-L", "--llm-base-url":
					p.args.LLMBaseURL = args[i+1]
				case "-M", "--llm-name":
					p.args.LLMName = args[i+1]
				case "-m", "--msg":
					p.args.Msg = args[i+1]
				}
				i++
			default:
				p.args.TransportArgs = append(p.args.TransportArgs, arg)
			}
		}
	}

	if err := p.checkArgs(); err != nil {
		return fmt.Errorf("check args: %w", err)
	}
	return nil
}

func (p *Parser) Arguments() Arguments {
	return p.args
}

func (p Parser) ParseData(arg string) (string, error) {
	after, ok := strings.CutPrefix(arg, "@")
	if !ok {
		return after, nil
	}
	d, err := os.ReadFile(after)
	if err != nil {
		return "", fmt.Errorf("read data file: %w", err)
	}
	return strings.TrimSpace(string(d)), nil
}

func (p Parser) ParseHeader(header string) ([]string, error) {
	var ret []string
	after, ok := strings.CutPrefix(header, "@")
	if !ok {
		ret = append(ret, after)
		return ret, nil
	}
	file, err := os.Open(after)
	if err != nil {
		return nil, fmt.Errorf("read header file: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if h := strings.TrimSpace(scanner.Text()); h != "" {
			ret = append(ret, h)
		}
	}
	return ret, nil
}

func (p *Parser) applyFromEnv() error {
	if v := os.Getenv("MCPURL_LLM_API_KEY"); v != "" {
		p.args.LLMApiKey = v
	}
	if v := os.Getenv("MCPURL_LLM_BASE_URL"); v != "" {
		p.args.LLMBaseURL = v
	}
	if v := os.Getenv("MCPURL_LLM_NAME"); v != "" {
		p.args.LLMName = v
	}
	if v := os.Getenv("MCPURL_LOG_LEVEL"); v != "" {
		if err := p.args.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return fmt.Errorf("parse log level: %w", err)
		}
	}
	if v := os.Getenv("MCPURL_HISTORY_FILE"); v != "" {
		p.args.HistoryFile = v
	} else {
		p.args.HistoryFile = historyFile()
	}
	if v := os.Getenv("MCPURL_LLM_CONTEXT_FILE"); v != "" {
		p.args.LLMContextFile = v
	} else {
		p.args.LLMContextFile = llmContextFile()
	}
	return nil
}

func (p Parser) checkArgs() error {
	if p.args.LLMBaseURL != "" && p.args.LLMName == "" {
		return fmt.Errorf("model name is required when LLM base url is set")
	}
	return nil
}

func historyFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".mcpurl_history")
}

func llmContextFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".mcpurl_llm_contexts")
}
