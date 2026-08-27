package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// CodexUsageURL is the ChatGPT backend Codex usage endpoint.
	CodexUsageURL = "https://chatgpt.com/backend-api/codex/usage"
	// WhamUsageURL is the ChatGPT WHAM usage endpoint (compatible payload).
	WhamUsageURL = "https://chatgpt.com/backend-api/wham/usage"

	// DefaultUserAgent is a browser-like UA; bare agents may get Cloudflare 403.
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
)

// GetOpts configures a backend GET.
type GetOpts struct {
	// URL is required.
	URL string
	// Auth credentials (AccessToken + AccountID required).
	Auth Auth
	// HTTPClient nil → http.DefaultClient.
	HTTPClient *http.Client
	// UserAgent empty → DefaultUserAgent.
	UserAgent string
	// ExtraHeaders merged after standard auth headers (may override).
	ExtraHeaders map[string]string
}

// GetJSON GETs url with Codex ChatGPT auth headers and returns the response body.
// Non-2xx responses return an error that includes the status code (body omitted
// from the error string to avoid leaking HTML/token material).
func GetJSON(ctx context.Context, opts GetOpts) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	url := strings.TrimSpace(opts.URL)
	if url == "" {
		return nil, fmt.Errorf("codex api: empty url")
	}
	if strings.TrimSpace(opts.Auth.AccessToken) == "" {
		return nil, fmt.Errorf("codex api: missing access_token")
	}
	if strings.TrimSpace(opts.Auth.AccountID) == "" {
		return nil, fmt.Errorf("codex api: missing account_id")
	}

	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	ua := strings.TrimSpace(opts.UserAgent)
	if ua == "" {
		ua = DefaultUserAgent
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("codex api: new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+opts.Auth.AccessToken)
	req.Header.Set("ChatGPT-Account-Id", opts.Auth.AccountID)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Origin", "https://chatgpt.com")
	req.Header.Set("Referer", "https://chatgpt.com/codex/settings/usage")
	for k, v := range opts.ExtraHeaders {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("codex api: get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("codex api: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("codex api: HTTP %d from usage endpoint", resp.StatusCode)
	}
	return body, nil
}

// GetCodexUsage GETs CodexUsageURL.
func GetCodexUsage(ctx context.Context, opts GetOpts) ([]byte, error) {
	opts.URL = CodexUsageURL
	return GetJSON(ctx, opts)
}

// GetWhamUsage GETs WhamUsageURL.
func GetWhamUsage(ctx context.Context, opts GetOpts) ([]byte, error) {
	opts.URL = WhamUsageURL
	return GetJSON(ctx, opts)
}
