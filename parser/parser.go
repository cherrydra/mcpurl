package parser

import (
	"errors"
	"slices"
)

var (
	ErrInvalidUsage = errors.New("invalid usage")
)

type Parser struct {
	Headers []string
	Help    bool
	Version bool

	transportArgs []string

	tools     bool
	prompts   bool
	resources bool

	tool     string
	data     string
	prompt   string
	resource string
}

func (p *Parser) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-T", "--tools":
			p.tools = true
		case "-P", "--prompts":
			p.prompts = true
		case "-R", "--resources":
			p.resources = true
		case "-h", "--help":
			p.Help = true
		case "-v", "--version":
			p.Version = true
		default:
			switch arg {
			case "-t", "--tool", "-p", "--prompt", "-r", "--resource", "-d", "--data", "-H", "--header":
				if len(args) < i+2 {
					return ErrInvalidUsage
				}
				switch arg {
				case "-t", "--tool":
					p.tool = args[i+1]
				case "-p", "--prompt":
					p.prompt = args[i+1]
				case "-r", "--resource":
					p.resource = args[i+1]
				case "-d", "--data":
					p.data = args[i+1]
				case "-H", "--header":
					p.Headers = append(p.Headers, args[i+1])
				}
				i++
			default:
				p.transportArgs = append(p.transportArgs, arg)
			}
		}
	}
	return nil
}

func (p *Parser) TransportArgs() []string {
	return slices.Clone(p.transportArgs)
}

func (p *Parser) Tools() bool {
	return p.tools
}

func (p *Parser) Prompts() bool {
	return p.prompts
}

func (p *Parser) Resources() bool {
	return p.resources
}

func (p *Parser) Tool() string {
	return p.tool
}

func (p *Parser) Data() string {
	return p.data
}

func (p *Parser) Prompt() string {
	return p.prompt
}

func (p *Parser) Resource() string {
	return p.resource
}
