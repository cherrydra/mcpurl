package features

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerFeatures struct {
	Session *mcp.ClientSession
	Out     *os.File
}

func (s ServerFeatures) CallTool(ctx context.Context, tool, data string) error {
	params := map[string]any{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &params); err != nil {
			return fmt.Errorf("unmarshal tool arguments: %w", err)
		}
	}
	return s.CallTool1(ctx, tool, params)
}

func (s ServerFeatures) CallTool1(ctx context.Context, tool string, params map[string]any) error {
	if s.Session == nil {
		return ErrNoSession
	}
	result, err := s.Session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tool,
		Arguments: params,
	})
	if err != nil {
		return fmt.Errorf("call tool: %w", err)
	}

	for _, c := range result.Content {
		out, _ := c.MarshalJSON()
		fmt.Fprintln(cmp.Or(s.Out, os.Stdout), string(out))
	}
	return nil
}

func (s ServerFeatures) CallTool2(ctx context.Context, tool string, arguments string) (mcp.Content, error) {
	if s.Session == nil {
		return nil, ErrNoSession
	}
	params := map[string]any{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return nil, fmt.Errorf("unmarshal tool arguments: %w", err)
		}
	}
	result, err := s.Session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tool,
		Arguments: params,
	})
	if err != nil {
		return nil, fmt.Errorf("call tool: %w", err)
	}

	return result.Content[0], nil
}

func (s ServerFeatures) GetPrompt(ctx context.Context, prompt, data string) error {
	params := map[string]string{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &params); err != nil {
			return fmt.Errorf("unmarshal prompt arguments: %w", err)
		}
	}
	return s.GetPrompt1(ctx, prompt, params)
}

func (s ServerFeatures) GetPrompt1(ctx context.Context, prompt string, params map[string]string) error {
	if s.Session == nil {
		return ErrNoSession
	}
	result, err := s.Session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      prompt,
		Arguments: params,
	})
	if err != nil {
		return fmt.Errorf("get prompt: %w", err)
	}

	for _, c := range result.Messages {
		json.NewEncoder(cmp.Or(s.Out, os.Stdout)).Encode(c)
	}
	return nil
}

func (s ServerFeatures) ListPrompts(ctx context.Context) ([]*mcp.Prompt, error) {
	if s.Session == nil {
		return nil, ErrNoSession
	}
	params := &mcp.ListPromptsParams{}
	var prompts []*mcp.Prompt
	for {
		result, err := s.Session.ListPrompts(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list prompts: %w", err)
		}
		prompts = append(prompts, result.Prompts...)
		if result.NextCursor == "" {
			break
		}
		params.Cursor = result.NextCursor
	}
	return prompts, nil
}

func (s ServerFeatures) PrintPrompts(ctx context.Context) error {
	prompts, err := s.ListPrompts(ctx)
	if err != nil {
		return err
	}
	for _, prompt := range prompts {
		json.NewEncoder(cmp.Or(s.Out, os.Stdout)).Encode(prompt)
	}
	return nil
}

func (s ServerFeatures) ListResources(ctx context.Context) ([]*mcp.Resource, error) {
	if s.Session == nil {
		return nil, ErrNoSession
	}
	params := &mcp.ListResourcesParams{}
	var resources []*mcp.Resource
	for {
		result, err := s.Session.ListResources(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list resources: %w", err)
		}
		resources = append(resources, result.Resources...)
		if result.NextCursor == "" {
			break
		}
		params.Cursor = result.NextCursor
	}
	return resources, nil
}

func (s ServerFeatures) PrintResources(ctx context.Context) error {
	resources, err := s.ListResources(ctx)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		json.NewEncoder(cmp.Or(s.Out, os.Stdout)).Encode(resource)
	}
	return nil
}

func (s ServerFeatures) ReadResource(ctx context.Context, resource string) error {
	if s.Session == nil {
		return ErrNoSession
	}
	result, err := s.Session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: resource,
	})
	if err != nil {
		return fmt.Errorf("read resource: %w", err)
	}
	for _, c := range result.Contents {
		json.NewEncoder(cmp.Or(s.Out, os.Stdout)).Encode(c)
	}
	return nil
}

func (s ServerFeatures) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	if s.Session == nil {
		return nil, ErrNoSession
	}
	params := &mcp.ListToolsParams{}
	var tools []*mcp.Tool
	for {
		result, err := s.Session.ListTools(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list tools: %w", err)
		}
		tools = append(tools, result.Tools...)
		if result.NextCursor == "" {
			break
		}
		params.Cursor = result.NextCursor
	}
	return tools, nil
}

func (s ServerFeatures) PrintTools(ctx context.Context) error {
	tools, err := s.ListTools(ctx)
	if err != nil {
		return err
	}
	for _, tool := range tools {
		json.NewEncoder(cmp.Or(s.Out, os.Stdout)).Encode(tool)
	}
	return nil
}
