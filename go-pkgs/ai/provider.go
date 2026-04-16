package ai

import (
	"context"

	"github.com/xhd2015/dot-pkgs/go-pkgs/ai/deepseek"
	"github.com/xhd2015/dot-pkgs/go-pkgs/ai/openai"
	aitypes "github.com/xhd2015/dot-pkgs/go-pkgs/ai/types"
)

// Provider defines the interface for AI providers
type Provider interface {
	// CallCompletion sends messages and returns the response (simple text only)
	CallCompletion(ctx context.Context, messages []aitypes.Message) (string, error)
	// CallStream sends messages and streams the response via callback
	CallStream(ctx context.Context, messages []aitypes.Message, callback aitypes.StreamCallback) error
	// CallWithTools sends messages with tool definitions and returns the response which may contain tool calls
	CallWithTools(ctx context.Context, messages []aitypes.ToolMessage, opts aitypes.ChatOptions) (*aitypes.ChatResponse, error)
}

// GetProvider returns the appropriate AI provider based on configuration
func GetProvider(cfg aitypes.Config) Provider {
	switch cfg.Provider {
	case aitypes.ProviderDeepSeek:
		return &deepseekProvider{cfg: cfg}
	case aitypes.ProviderMoonshot:
		return &openaiProvider{cfg: cfg}
	default:
		return &openaiProvider{cfg: cfg}
	}
}

// openaiProvider implements Provider using the OpenAI SDK
type openaiProvider struct {
	cfg aitypes.Config
}

func (p *openaiProvider) CallCompletion(ctx context.Context, messages []aitypes.Message) (string, error) {
	return openai.CallCompletion(ctx, p.cfg, messages)
}

func (p *openaiProvider) CallStream(ctx context.Context, messages []aitypes.Message, callback aitypes.StreamCallback) error {
	return openai.CallStream(ctx, p.cfg, messages, callback)
}

func (p *openaiProvider) CallWithTools(ctx context.Context, messages []aitypes.ToolMessage, opts aitypes.ChatOptions) (*aitypes.ChatResponse, error) {
	return openai.CallWithTools(ctx, p.cfg, messages, opts)
}

// deepseekProvider implements Provider using custom HTTP for DeepSeek
type deepseekProvider struct {
	cfg aitypes.Config
}

func (p *deepseekProvider) CallCompletion(ctx context.Context, messages []aitypes.Message) (string, error) {
	return deepseek.CallCompletion(ctx, p.cfg, messages)
}

func (p *deepseekProvider) CallStream(ctx context.Context, messages []aitypes.Message, callback aitypes.StreamCallback) error {
	return deepseek.CallStream(ctx, p.cfg, messages, callback)
}

func (p *deepseekProvider) CallWithTools(ctx context.Context, messages []aitypes.ToolMessage, opts aitypes.ChatOptions) (*aitypes.ChatResponse, error) {
	return deepseek.CallWithTools(ctx, p.cfg, messages, opts)
}
