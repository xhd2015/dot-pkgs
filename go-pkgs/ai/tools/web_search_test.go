package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/ai"
	aitypes "github.com/xhd2015/dot-pkgs/go-pkgs/ai/types"
	"github.com/xhd2015/dot-pkgs/go-pkgs/testconfig"
)

func TestDeepSeekWithWebSearch(t *testing.T) {
	tc := testconfig.RequireAI(t)
	tcSearch := testconfig.RequireWebSearch(t)

	searchCfg := WebSearchConfig{
		BochaAPIKey:     tcSearch.BochaAPIKey,
		SerpAPIKey:      tcSearch.SerpAPIKey,
		GoogleSearchKey: tcSearch.GoogleSearchKey,
		GoogleSearchCX:  tcSearch.GoogleSearchCX,
	}

	providerName := aitypes.Provider(tc.AIProvider)
	if providerName == "" {
		providerName = aitypes.ProviderDeepSeek
	}

	cfg := aitypes.Config{
		Provider: providerName,
		APIKey:   tc.AIAPIKey,
		BaseURL:  tc.AIBaseURL,
		Model:    tc.AIModel,
	}

	provider := ai.GetProvider(cfg)

	toolDef := WebSearchToolDefinition()
	messages := []aitypes.ToolMessage{
		{Message: aitypes.Message{Role: "system", Content: "You are a travel assistant. Use the web_search tool when you need up-to-date information."}},
		{Message: aitypes.Message{Role: "user", Content: "What are the current visa requirements for Chinese citizens traveling to Japan in 2025?"}},
	}
	opts := aitypes.ChatOptions{Tools: []aitypes.Tool{toolDef}}

	ctx := context.Background()

	const maxRounds = 5
	for round := 0; round < maxRounds; round++ {
		resp, err := provider.CallWithTools(ctx, messages, opts)
		if err != nil {
			t.Fatalf("round %d: CallWithTools failed: %v", round, err)
		}

		t.Logf("round %d: content=%q, tool_calls=%d", round, truncate(resp.Content, 100), len(resp.ToolCalls))

		if len(resp.ToolCalls) == 0 {
			if resp.Content == "" {
				t.Fatal("Got empty response with no tool calls")
			}
			t.Logf("Final answer: %s", truncate(resp.Content, 500))
			return
		}

		messages = append(messages, aitypes.ToolMessage{
			Message:   aitypes.Message{Role: "assistant", Content: resp.Content},
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			t.Logf("Tool call: %s(%s)", tc.Name, truncate(tc.Arguments, 200))

			if tc.Name != "web_search" {
				t.Fatalf("Unexpected tool call: %s", tc.Name)
			}

			var args struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
				t.Fatalf("Failed to parse tool call args: %v", err)
			}

			result, err := WebSearch(ctx, searchCfg, args.Query)
			if err != nil {
				result = fmt.Sprintf("Search failed: %v", err)
			}
			t.Logf("Search result: %s", truncate(result, 300))

			messages = append(messages, aitypes.ToolMessage{
				Message: aitypes.Message{Role: "tool"},
				ToolResult: &aitypes.ToolResult{
					ToolCallID: tc.ID,
					Name:       tc.Name,
					Content:    result,
				},
			})
		}
	}

	t.Fatal("Exceeded max tool call rounds without getting a final answer")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
