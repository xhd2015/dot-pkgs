package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	aitypes "github.com/xhd2015/dot-pkgs/go-pkgs/ai/types"
)

// WebSearchConfig holds configuration for the web search tool.
// Priority: Bocha > SerpAPI > Google Custom Search.
type WebSearchConfig struct {
	BochaAPIKey     string
	SerpAPIKey      string
	GoogleSearchKey string
	GoogleSearchCX  string
}

func (c *WebSearchConfig) Enabled() bool {
	return c.BochaAPIKey != "" || c.SerpAPIKey != "" || c.GoogleSearchKey != ""
}

// WebSearchToolDefinition returns the AI tool definition for web search
func WebSearchToolDefinition() aitypes.Tool {
	return aitypes.Tool{
		Name:        "web_search",
		Description: "Search the web for up-to-date information about travel destinations, visa policies, prices, opening hours, reviews, weather, etc.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search query. Use the language most likely to yield good results.",
				},
			},
			"required": []string{"query"},
		},
	}
}

// WebSearch performs a web search using the configured search provider
func WebSearch(ctx context.Context, cfg WebSearchConfig, query string) (string, error) {
	if cfg.BochaAPIKey != "" {
		return searchBocha(ctx, cfg.BochaAPIKey, query)
	}
	if cfg.SerpAPIKey != "" {
		return searchSerpAPI(ctx, cfg.SerpAPIKey, query)
	}
	if cfg.GoogleSearchKey != "" {
		return searchGoogle(ctx, cfg.GoogleSearchKey, cfg.GoogleSearchCX, query)
	}
	return "", fmt.Errorf("no web search API configured (set BochaAPIKey, SerpAPIKey, or GoogleSearchKey)")
}

func searchBocha(ctx context.Context, apiKey string, query string) (string, error) {
	reqBody, err := json.Marshal(map[string]any{
		"query":   query,
		"summary": true,
		"count":   5,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.bochaai.com/v1/web-search", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Bocha API error (status %d): %s", resp.StatusCode, truncateBytes(body, 200))
	}

	var result struct {
		Data struct {
			WebPages struct {
				Value []struct {
					Name    string `json:"name"`
					URL     string `json:"url"`
					Summary string `json:"summary"`
				} `json:"value"`
			} `json:"webPages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse Bocha response: %w", err)
	}

	pages := result.Data.WebPages.Value
	if len(pages) == 0 {
		return "No results found.", nil
	}

	var sb strings.Builder
	for i, p := range pages {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n   %s\n", i+1, p.Name, p.Summary, p.URL))
	}
	return sb.String(), nil
}

func searchSerpAPI(ctx context.Context, apiKey string, query string) (string, error) {
	apiURL := fmt.Sprintf("https://serpapi.com/search.json?q=%s&api_key=%s&num=5",
		url.QueryEscape(query), apiKey)

	body, err := httpGet(ctx, apiURL)
	if err != nil {
		return "", err
	}

	var result struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
		AnswerBox *struct {
			Answer  string `json:"answer"`
			Snippet string `json:"snippet"`
		} `json:"answer_box"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	var sb strings.Builder
	if result.AnswerBox != nil {
		answer := result.AnswerBox.Answer
		if answer == "" {
			answer = result.AnswerBox.Snippet
		}
		if answer != "" {
			sb.WriteString(fmt.Sprintf("Quick answer: %s\n\n", answer))
		}
	}

	for i, r := range result.OrganicResults {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n   %s\n", i+1, r.Title, r.Snippet, r.Link))
	}

	if sb.Len() == 0 {
		return "No results found.", nil
	}
	return sb.String(), nil
}

func searchGoogle(ctx context.Context, apiKey string, cx string, query string) (string, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?q=%s&key=%s&cx=%s&num=5",
		url.QueryEscape(query), apiKey, cx)

	body, err := httpGet(ctx, apiURL)
	if err != nil {
		return "", err
	}

	var result struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "No results found.", nil
	}

	var sb strings.Builder
	for i, item := range result.Items {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n   %s\n", i+1, item.Title, item.Snippet, item.Link))
	}
	return sb.String(), nil
}

func httpGet(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func truncateBytes(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
