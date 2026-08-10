package eventbus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Publisher posts Event JSON to a hub's HTTP publish endpoint.
// An empty base URL makes Publish a no-op success (disabled publisher).
type Publisher struct {
	baseURL string
	token   string
	client  *http.Client
	timeout time.Duration
}

// Option configures a Publisher.
type Option func(*Publisher)

// WithToken sets the optional Bearer token for Authorization.
// Empty token means no Authorization header is sent.
func WithToken(token string) Option {
	return func(p *Publisher) {
		p.token = token
	}
}

// WithTimeout sets the HTTP client timeout used when no custom client is provided.
func WithTimeout(d time.Duration) Option {
	return func(p *Publisher) {
		p.timeout = d
	}
}

// WithHTTPClient injects an HTTP client (for tests and custom transports).
// When set, WithTimeout does not replace the client; callers control timeouts.
func WithHTTPClient(c *http.Client) Option {
	return func(p *Publisher) {
		p.client = c
	}
}

// NewPublisher builds a Publisher for baseURL (e.g. "http://127.0.0.1:23891").
// Trailing slashes on baseURL are stripped. Empty baseURL is allowed (no-op Publish).
func NewPublisher(baseURL string, opts ...Option) *Publisher {
	p := &Publisher{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	if p.client == nil {
		c := &http.Client{}
		if p.timeout > 0 {
			c.Timeout = p.timeout
		}
		p.client = c
	}
	return p
}

// Publish POSTs the Event as JSON to {baseURL}/publish.
// Empty base URL returns nil without performing HTTP.
// Non-2xx responses and transport failures return an error.
func (p *Publisher) Publish(ctx context.Context, ev Event) error {
	if p == nil || p.baseURL == "" {
		return nil
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	url := p.baseURL + "/publish"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	client := p.client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("eventbus: publish %s: status %d", url, resp.StatusCode)
	}
	return nil
}
