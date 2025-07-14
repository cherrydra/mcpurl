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

	"github.com/cherrydra/mcpurl/mcp/features"
	"github.com/openai/openai-go"
)

var (
	ErrDisabled = errors.New("llm disabled")
)

type LLM struct {
	Client *openai.Client
	Model  string

	ContextManger TalkContextManager
}

func (i *LLM) Msg(ctx context.Context, f features.ServerFeatures, message string, out *os.File) error {
	if i.Client == nil {
		return ErrDisabled
	}
	tools, err := f.ListTools(ctx)
	if err != nil && !errors.Is(err, features.ErrNoSession) {
		return fmt.Errorf("list tools: %w", err)
	}
	params := openai.ChatCompletionNewParams{
		Messages: i.ContextManger.addMsg(openai.UserMessage(message)).Messages,
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
		if detector.TotalBytes() > 0 && detector.LastByte() != '\n' {
			fmt.Fprintln(out)
		}
		if stream.Err() != nil {
			if errors.Is(stream.Err(), context.Canceled) {
				return context.Canceled
			}
			return fmt.Errorf("streaming error: %w", stream.Err())
		}
		if len(acc.Choices) == 0 {
			return errors.New("no choices in response")
		}
		params.Messages = append(params.Messages, acc.Choices[0].Message.ToParam())
		switch acc.Choices[0].FinishReason {
		case "tool_calls":
			if len(acc.Choices[0].Message.ToolCalls) == 0 {
				return errors.New("no tool calls in response, but finish reason is tool_calls")
			}
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
			i.ContextManger.setMsgs(params.Messages)
			return nil
		case "length":
			return errors.New("response too long for model")
		default:
			return fmt.Errorf("unexpected finish reason: %s", acc.Choices[0].FinishReason)
		}
	}
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
