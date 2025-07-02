package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func callTool(ctx context.Context, session *mcp.ClientSession, tool, data string) error {
	params := map[string]any{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &params); err != nil {
			return fmt.Errorf("unmarshal tool arguments: %w", err)
		}
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tool,
		Arguments: params,
	})
	if err != nil {
		return fmt.Errorf("call tool: %w", err)
	}

	for _, c := range result.Content {
		out, _ := c.MarshalJSON()
		fmt.Println(string(out))
	}
	return nil
}
