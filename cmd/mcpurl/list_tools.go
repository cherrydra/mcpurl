package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func listTools(ctx context.Context, session *mcp.ClientSession) error {
	params := &mcp.ListToolsParams{}
	for {
		result, err := session.ListTools(ctx, params)
		if err != nil {
			return fmt.Errorf("list tools: %w", err)
		}
		for _, tool := range result.Tools {
			json.NewEncoder(os.Stdout).Encode(tool)
		}
		if result.NextCursor == "" {
			break
		}
		params.Cursor = result.NextCursor
	}
	return nil
}
