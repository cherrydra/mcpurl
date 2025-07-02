package parser

import (
	"errors"
	"slices"
)

var (
	ErrInvalidUsage = errors.New("invalid usage")
)

type Parser struct {
	Help bool

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
	parseSubCommand := func() error {
		for i, arg := range args {
			switch arg {
			case "-t", "-p", "-r", "--tools", "--prompts", "--resources":
				p.transportArgs = append(p.transportArgs, args[:i]...)
				p.transportArgs = append(p.transportArgs, args[i+1:]...)
				switch arg {
				case "-t", "--tools":
					p.tools = true
				case "-p", "--prompts":
					p.prompts = true
				case "-r", "--resources":
					p.resources = true
				}
				return nil
			case "-T", "-P", "-R", "--tool", "--prompt", "--resource":
				if len(args) < i+2 {
					return ErrInvalidUsage
				}
				p.transportArgs = append(p.transportArgs, args[:i]...)
				if len(args) > i+2 {
					p.transportArgs = append(p.transportArgs, args[i+2:]...)
				}
				switch arg {
				case "-T", "--tool":
					p.tool = args[i+1]
				case "-P", "--prompt":
					p.prompt = args[i+1]
				case "-R", "--resource":
					p.resource = args[i+1]
				}
				return nil
			case "-h", "--help":
				p.Help = true
				return nil
			}
		}
		return ErrInvalidUsage
	}
	if err := parseSubCommand(); err != nil {
		return err
	}
	parseData := func() error {
		args := slices.Clone(p.transportArgs)
		for i, arg := range args {
			switch arg {
			case "-d", "--data":
				if len(args) < i+2 {
					return ErrInvalidUsage
				}
				p.transportArgs = append(p.transportArgs[:0], args[:i]...)
				if len(args) > i+2 {
					p.transportArgs = append(p.transportArgs, args[i+2:]...)
				}
				p.data = args[i+1]
				return nil
			}
		}
		return nil
	}
	if err := parseData(); err != nil {
		return err
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
