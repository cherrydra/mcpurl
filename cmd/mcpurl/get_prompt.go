package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func getPrompt(ctx context.Context, session *mcp.ClientSession, prompt, data string) error {
	params := map[string]string{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &params); err != nil {
			return fmt.Errorf("unmarshal prompt arguments: %w", err)
		}
	}
	result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      prompt,
		Arguments: params,
	})
	if err != nil {
		return fmt.Errorf("get prompt: %w", err)
	}

	for _, c := range result.Messages {
		json.NewEncoder(os.Stdout).Encode(c)
	}
	return nil
}
