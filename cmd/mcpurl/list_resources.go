package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func listResources(ctx context.Context, session *mcp.ClientSession) error {
	params := &mcp.ListResourcesParams{}
	for {
		result, err := session.ListResources(ctx, params)
		if err != nil {
			return fmt.Errorf("list resources: %w", err)
		}
		for _, resource := range result.Resources {
			json.NewEncoder(os.Stdout).Encode(resource)
		}
		if result.NextCursor == "" {
			break
		}
		params.Cursor = result.NextCursor
	}
	return nil
}
