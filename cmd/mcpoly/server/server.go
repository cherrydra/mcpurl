package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/cherrydra/mcpurl/cmd/mcpoly/config"
	"github.com/cherrydra/mcpurl/mcp/transport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ReverseProxy struct {
	s        *mcp.Server
	c        *mcp.Client
	css      map[string]*mcp.ClientSession
	initOnce sync.Once
}

func (s *ReverseProxy) AddBackends(ctx context.Context, servers map[string]config.Server) error {
	s.initOnce.Do(func() {
		s.s = mcp.NewServer(&mcp.Implementation{}, nil)
		s.c = mcp.NewClient(&mcp.Implementation{}, nil)
		s.css = make(map[string]*mcp.ClientSession)
	})
	for k, v := range servers {
		var cs *mcp.ClientSession
		var err error
		switch v.Type {
		case "stdio":
			cmd := exec.Command(v.Command, v.Args...)
			cmd.Env = v.Env.Encode()
			cmd.Stderr = os.Stderr
			cs, err = s.c.Connect(ctx, mcp.NewCommandTransport(cmd))
		case "http":
			cs, err = s.c.Connect(ctx, mcp.NewStreamableClientTransport(v.URL, &mcp.StreamableClientTransportOptions{
				HTTPClient: &http.Client{Transport: &transport.AddHeadersRoundTripper{Headers: v.Headers.Encode()}},
			}))
		case "sse":
			cs, err = s.c.Connect(ctx, mcp.NewSSEClientTransport(v.URL, &mcp.SSEClientTransportOptions{
				HTTPClient: &http.Client{Transport: &transport.AddHeadersRoundTripper{Headers: v.Headers.Encode()}},
			}))
		default:
			err = errors.New("unsupported server type: " + v.Type)
		}
		if err != nil {
			return fmt.Errorf("failed to connect to server %s: %w", k, err)
		}

		s.css[k] = cs
		s.addTools(ctx, cs)
		s.addPrompts(ctx, cs)
		s.addResources(ctx, cs)
	}
	return nil
}

func (s *ReverseProxy) Run(ctx context.Context) error {
	if s.s == nil {
		return errors.New("no backends added")
	}
	err := s.s.Run(ctx, mcp.NewStdioTransport())
	for k, v := range s.css {
		if err := v.Close(); err != nil {
			slog.Error("close client session", "server", k, "err", err)
		}
	}
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (s *ReverseProxy) addTools(ctx context.Context, cs *mcp.ClientSession) {
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			if strings.Contains(err.Error(), "Method not found") {
				break
			}
			slog.Error("list tools", "err", err)
			continue
		}
		tool.InputSchema.Schema = ""
		s.s.AddTool(tool, func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[any], error) {
			return cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      params.Name,
				Arguments: params.Arguments,
			})
		})
	}
}

func (s *ReverseProxy) addPrompts(ctx context.Context, cs *mcp.ClientSession) {
	for prompt, err := range cs.Prompts(ctx, nil) {
		if err != nil {
			if strings.Contains(err.Error(), "Method not found") {
				break
			}
			slog.Error("list prompts", "err", err)
			continue
		}
		s.s.AddPrompt(prompt, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
			return cs.GetPrompt(ctx, params)
		})
	}
}

func (s *ReverseProxy) addResources(ctx context.Context, cs *mcp.ClientSession) {
	for resource, err := range cs.Resources(ctx, nil) {
		if err != nil {
			if strings.Contains(err.Error(), "Method not found") {
				break
			}
			slog.Error("list resources", "err", err)
			continue
		}
		s.s.AddResource(resource, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
			return cs.ReadResource(ctx, params)
		})
	}
}
