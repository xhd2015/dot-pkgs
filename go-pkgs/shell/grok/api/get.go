package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// BillingURL is the cli-chat-proxy billing/usage endpoint.
	BillingURL = "https://cli-chat-proxy.grok.com/v1/billing"

	// DefaultUserAgent identifies Grok CLI-style clients to the proxy.
	DefaultUserAgent = "GrokCLI/1.0.5"
)

// GetOpts configures a backend GET.
type GetOpts struct {
	// URL is required.
	URL string
	// Auth credentials (AccessToken required).
	Auth Auth
	// HTTPClient nil → http.DefaultClient.
	HTTPClient *http.Client
	// UserAgent empty → DefaultUserAgent.
	UserAgent string
	// ExtraHeaders merged after standard auth headers (may override).
	ExtraHeaders map[string]string
}

// GetJSON GETs url with Bearer auth and returns the response body.
// Non-2xx responses return an error that includes the status code (body omitted
// from the error string to avoid leaking HTML/token material).
func GetJSON(ctx context.Context, opts GetOpts) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	url := strings.TrimSpace(opts.URL)
	if url == "" {
		return nil, fmt.Errorf("grok api: empty url")
	}
	if strings.TrimSpace(opts.Auth.AccessToken) == "" {
		return nil, fmt.Errorf("grok api: missing access token")
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
		return nil, fmt.Errorf("grok api: new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+opts.Auth.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", ua)
	for k, v := range opts.ExtraHeaders {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("grok api: get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("grok api: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("grok api: HTTP %d from %s", resp.StatusCode, shortURL(url))
	}
	return body, nil
}

// GetBilling GETs BillingURL.
func GetBilling(ctx context.Context, opts GetOpts) ([]byte, error) {
	opts.URL = BillingURL
	return GetJSON(ctx, opts)
}

// IsUnauthorized reports whether err looks like an HTTP 401/403 from GetJSON.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTTP 401") || strings.Contains(msg, "HTTP 403")
}

func shortURL(u string) string {
	// Keep path hint without query; avoid dumping huge URLs into errors.
	if i := strings.Index(u, "?"); i >= 0 {
		u = u[:i]
	}
	if len(u) > 120 {
		return u[:117] + "..."
	}
	return u
}
