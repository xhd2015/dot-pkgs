package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	aitypes "github.com/xhd2015/dot-pkgs/go-pkgs/ai/types"
)

// Default DeepSeek API endpoint
const DefaultBaseURL = "https://api.deepseek.com"

// streamRequest represents DeepSeek's streaming request
type streamRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// streamChoice represents a choice in DeepSeek's streaming response
type streamChoice struct {
	Delta struct {
		Content          string `json:"content"`
		ReasoningContent string `json:"reasoning_content"` // DeepSeek R1 thinking
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

// streamResponse represents DeepSeek's streaming response chunk
type streamResponse struct {
	Choices []streamChoice `json:"choices"`
	Usage   *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// CallCompletion calls DeepSeek API for non-streaming completion
func CallCompletion(ctx context.Context, cfg aitypes.Config, messages []aitypes.Message) (string, error) {
	// Convert messages
	reqMessages := make([]message, len(messages))
	for i, msg := range messages {
		reqMessages[i] = message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	model := cfg.Model
	if model == "" {
		model = "deepseek-chat"
	}

	reqBody := streamRequest{
		Model:    model,
		Stream:   false,
		Messages: reqMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	url := baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API error (model: %s, status %d): %s", model, resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("AI error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, nil
}

// CallWithTools calls DeepSeek API with tool support.
// DeepSeek tool calling uses the same format as OpenAI, so we reuse the completion endpoint.
func CallWithTools(ctx context.Context, cfg aitypes.Config, messages []aitypes.ToolMessage, opts aitypes.ChatOptions) (*aitypes.ChatResponse, error) {
	// Convert tool messages to simple messages (DeepSeek doesn't natively support tool calling in this implementation)
	simple := make([]aitypes.Message, 0, len(messages))
	for _, m := range messages {
		if m.ToolResult != nil {
			simple = append(simple, aitypes.Message{Role: "user", Content: fmt.Sprintf("[Tool result for %s]: %s", m.ToolResult.Name, m.ToolResult.Content)})
			continue
		}
		simple = append(simple, m.Message)
	}
	content, err := CallCompletion(ctx, cfg, simple)
	if err != nil {
		return nil, err
	}
	return &aitypes.ChatResponse{Content: content}, nil
}

// CallStream calls DeepSeek API with streaming, capturing reasoning_content (thinking)
func CallStream(ctx context.Context, cfg aitypes.Config, messages []aitypes.Message, callback aitypes.StreamCallback) error {
	// Convert messages
	reqMessages := make([]message, len(messages))
	for i, msg := range messages {
		reqMessages[i] = message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	model := cfg.Model
	if model == "" {
		model = "deepseek-chat"
	}

	reqBody := streamRequest{
		Model:    model,
		Stream:   true,
		Messages: reqMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Construct the full URL (BaseURL + /chat/completions)
	url := baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI API error (model: %s, status %d): %s", model, resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	var lastUsage *aitypes.TokenUsage

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read stream: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// SSE format: "data: {json}"
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))
		if string(data) == "[DONE]" {
			callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeDone, Content: "", TokenUsage: lastUsage})
			break
		}

		var streamResp streamResponse
		if err := json.Unmarshal(data, &streamResp); err != nil {
			continue // Skip malformed chunks
		}

		if streamResp.Error != nil {
			return fmt.Errorf("AI error: %s", streamResp.Error.Message)
		}

		// Capture usage info if present (usually in the last chunk)
		if streamResp.Usage != nil {
			lastUsage = &aitypes.TokenUsage{
				PromptTokens:     streamResp.Usage.PromptTokens,
				CompletionTokens: streamResp.Usage.CompletionTokens,
				TotalTokens:      streamResp.Usage.TotalTokens,
			}
		}

		if len(streamResp.Choices) == 0 {
			continue
		}

		choice := streamResp.Choices[0]

		// Check for finish
		if choice.FinishReason == "stop" {
			callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeDone, Content: "", TokenUsage: lastUsage})
			break
		}

		// Send reasoning/thinking content (DeepSeek R1 specific)
		if choice.Delta.ReasoningContent != "" {
			if err := callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeThinking, Content: choice.Delta.ReasoningContent}); err != nil {
				return err
			}
		}

		// Send regular content
		if choice.Delta.Content != "" {
			if err := callback(aitypes.StreamChunk{Type: aitypes.ChunkTypeContent, Content: choice.Delta.Content}); err != nil {
				return err
			}
		}
	}

	return nil
}
