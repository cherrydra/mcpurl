package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func readResource(ctx context.Context, session *mcp.ClientSession, resource string) error {
	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: resource,
	})
	if err != nil {
		return fmt.Errorf("read resource: %w", err)
	}
	for _, c := range result.Contents {
		json.NewEncoder(os.Stdout).Encode(c)
	}
	return nil
}
