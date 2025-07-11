package llm

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/cherrydra/mcpurl/features"
	"github.com/openai/openai-go"
)

var (
	ErrDisabled = errors.New("llm disabled")
)

type LLM struct {
	Client *openai.Client
	Model  string

	messagesContextMutex sync.RWMutex
	messagesContext      []openai.ChatCompletionMessageParamUnion
}

func (i *LLM) Msg(ctx context.Context, f features.ServerFeatures, message string, out *os.File) error {
	i.messagesContextMutex.Lock()
	defer i.messagesContextMutex.Unlock()
	if i.Client == nil {
		return ErrDisabled
	}
	tools, err := f.ListTools(ctx)
	if err != nil && !errors.Is(err, features.ErrNoSession) {
		return fmt.Errorf("list tools: %w", err)
	}
	params := openai.ChatCompletionNewParams{
		Messages: append(i.messagesContext, openai.UserMessage(message)),
		Model:    i.Model,
	}

	for _, tool := range tools {
		schema, err := tool.InputSchema.MarshalJSON()
		if err != nil {
			return fmt.Errorf("marshal tool schema: %w", err)
		}
		var parameters openai.FunctionParameters
		if err := json.Unmarshal(schema, &parameters); err != nil {
			return fmt.Errorf("unmarshal tool schema: %w", err)
		}
		t := openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(cmp.Or(tool.Description, tool.Name)),
				Parameters:  parameters,
			},
		}
		params.Tools = append(params.Tools, t)
	}

	for {
		stream := i.Client.Chat.Completions.NewStreaming(ctx, params)
		acc := openai.ChatCompletionAccumulator{}
		detector := &LastByteDetector{}
		for stream.Next() {
			if stream.Err() != nil {
				return fmt.Errorf("streaming error: %w", stream.Err())
			}
			chunk := stream.Current()
			acc.AddChunk(chunk)
			fmt.Fprint(io.MultiWriter(out, detector), chunk.Choices[0].Delta.Content)
		}
		if len(acc.Choices) == 0 {
			return fmt.Errorf("no response from llm: %w", stream.Err())
		}
		if detector.TotalBytes() > 0 && detector.LastByte() != '\n' {
			fmt.Fprintln(out)
		}
		switch acc.Choices[0].FinishReason {
		case "tool_calls":
			if len(acc.Choices[0].Message.ToolCalls) == 0 {
				return errors.New("no tool calls in response, but finish reason is tool_calls")
			}
			params.Messages = append(params.Messages, acc.Choices[0].Message.ToParam())
			for _, toolCall := range acc.Choices[0].Message.ToolCalls {
				slog.Info("Calling tool", "name", toolCall.Function.Name, "arguments", toolCall.Function.Arguments)
				result, err := f.CallTool2(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
				if err != nil {
					return fmt.Errorf("call tool %s: %w", toolCall.Function.Name, err)
				}
				c, _ := result.MarshalJSON()
				slog.Info("Tool result", "name", toolCall.Function.Name, "result", string(c))
				params.Messages = append(params.Messages, openai.ToolMessage(string(c), toolCall.ID))
			}
		case "stop":
			fmt.Fprintln(out)
			i.messagesContext = append(i.messagesContext, acc.Choices[0].Message.ToParam())
			return nil
		case "length":
			return errors.New("response too long for model")
		default:
			return fmt.Errorf("unexpected finish reason: %s", acc.Choices[0].FinishReason)
		}
	}
}

func (i *LLM) ClearContext() {
	i.messagesContextMutex.Lock()
	defer i.messagesContextMutex.Unlock()
	i.messagesContext = []openai.ChatCompletionMessageParamUnion{}
}

type LastByteDetector struct {
	lastByte   byte
	totalBytes int64
}

func (d *LastByteDetector) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	n = len(p)
	d.lastByte = p[n-1]
	d.totalBytes += int64(n)
	return
}

func (d *LastByteDetector) LastByte() byte {
	return d.lastByte
}

func (d *LastByteDetector) TotalBytes() int64 {
	return d.totalBytes
}
