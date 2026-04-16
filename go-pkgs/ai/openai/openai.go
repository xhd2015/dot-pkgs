package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	openaisdk "github.com/sashabaranov/go-openai"
	aitypes "github.com/xhd2015/dot-pkgs/go-pkgs/ai/types"
)

// getClient creates an OpenAI client configured for the specified provider
func getClient(cfg aitypes.Config) *openaisdk.Client {
	clientCfg := openaisdk.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}
	return openaisdk.NewClientWithConfig(clientCfg)
}

// CallCompletion calls the AI API for a non-streaming completion
func CallCompletion(ctx context.Context, cfg aitypes.Config, messages []aitypes.Message) (string, error) {
	client := getClient(cfg)

	// Convert messages to OpenAI format
	openaiMessages := make([]openaisdk.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openaisdk.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	req := openaisdk.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
	}
	if cfg.MaxTokens > 0 {
		req.MaxTokens = cfg.MaxTokens
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("AI API error (model: %s): %w", model, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return resp.Choices[0].Message.Content, nil
}

// CallWithTools calls the AI API with tool/function calling support
func CallWithTools(ctx context.Context, cfg aitypes.Config, messages []aitypes.ToolMessage, opts aitypes.ChatOptions) (*aitypes.ChatResponse, error) {
	client := getClient(cfg)

	openaiMessages := convertToolMessages(messages)

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	req := openaisdk.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
	}
	if cfg.MaxTokens > 0 {
		req.MaxTokens = cfg.MaxTokens
	}

	if len(opts.Tools) > 0 {
		for _, t := range opts.Tools {
			req.Tools = append(req.Tools, openaisdk.Tool{
				Type: openaisdk.ToolTypeFunction,
				Function: &openaisdk.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			})
		}
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI API error (model: %s): %w", model, err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	choice := resp.Choices[0]
	result := &aitypes.ChatResponse{
		Content: choice.Message.Content,
	}

	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, aitypes.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}

func convertToolMessages(messages []aitypes.ToolMessage) []openaisdk.ChatCompletionMessage {
	result := make([]openaisdk.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		m := openaisdk.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		if msg.ToolResult != nil {
			m.Role = openaisdk.ChatMessageRoleTool
			m.Content = msg.ToolResult.Content
			m.Name = msg.ToolResult.Name
			m.ToolCallID = msg.ToolResult.ToolCallID
		}
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				m.ToolCalls = append(m.ToolCalls, openaisdk.ToolCall{
					ID:   tc.ID,
					Type: openaisdk.ToolTypeFunction,
					Function: openaisdk.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				})
			}
		}
		result = append(result, m)
	}
	return result
}

// CallStream calls the AI API with streaming enabled using the official SDK
func CallStream(ctx context.Context, cfg aitypes.Config, messages []aitypes.Message, callback aitypes.StreamCallback) error {
	client := getClient(cfg)

	// Convert messages to OpenAI format
	openaiMessages := make([]openaisdk.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openaisdk.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	streamReq := openaisdk.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
		Stream:   true,
	}
	if cfg.MaxTokens > 0 {
		streamReq.MaxTokens = cfg.MaxTokens
	}

	stream, err := client.CreateChatCompletionStream(ctx, streamReq)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeDone, Content: ""})
			return nil
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) == 0 {
			continue
		}

		choice := response.Choices[0]

		if choice.FinishReason == openaisdk.FinishReasonStop {
			callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeDone, Content: ""})
			return nil
		}

		content := choice.Delta.Content
		if content != "" {
			if err := callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeContent, Content: content}); err != nil {
				return err
			}
		}
	}
}
