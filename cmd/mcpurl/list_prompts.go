package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func listPrompts(ctx context.Context, session *mcp.ClientSession) error {
	params := &mcp.ListPromptsParams{}
	for {
		result, err := session.ListPrompts(ctx, params)
		if err != nil {
			return fmt.Errorf("list prompts: %w", err)
		}
		for _, prompt := range result.Prompts {
			json.NewEncoder(os.Stdout).Encode(prompt)
		}
		if result.NextCursor == "" {
			break
		}
		params.Cursor = result.NextCursor
	}
	return nil
}
