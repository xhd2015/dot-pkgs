// Package usage fetches and normalizes Codex account usage via shell/codex/api.
package usage

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/api"
)

// FetchOpts configures Fetch.
type FetchOpts struct {
	// AuthPath empty → api.DefaultAuthPath(). Ignored when Auth is set.
	AuthPath string
	// Auth when non-zero AccessToken skips file load (tests).
	Auth api.Auth
	// HTTPClient nil → http.DefaultClient (via api).
	HTTPClient *http.Client
	// UserAgent empty → api.DefaultUserAgent.
	UserAgent string
	// LoadAuth injectable; nil → api.LoadAuth.
	LoadAuth func(path string) (api.Auth, error)
	// GetJSON injectable; nil → api.GetJSON. When set, URL is still chosen by Fetch.
	GetJSON func(ctx context.Context, opts api.GetOpts) ([]byte, error)
}

// Fetch loads auth and GETs Codex usage (fallback WHAM), then normalizes.
func Fetch(ctx context.Context, opts FetchOpts) (Snapshot, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	auth := opts.Auth
	if strings.TrimSpace(auth.AccessToken) == "" {
		path := strings.TrimSpace(opts.AuthPath)
		if path == "" {
			var err error
			path, err = api.DefaultAuthPath()
			if err != nil {
				return Snapshot{}, err
			}
		}
		load := opts.LoadAuth
		if load == nil {
			load = api.LoadAuth
		}
		var err error
		auth, err = load(path)
		if err != nil {
			return Snapshot{}, err
		}
	}

	get := opts.GetJSON
	if get == nil {
		get = api.GetJSON
	}
	base := api.GetOpts{
		Auth:       auth,
		HTTPClient: opts.HTTPClient,
		UserAgent:  opts.UserAgent,
	}

	codexOpts := base
	codexOpts.URL = api.CodexUsageURL
	raw, err := get(ctx, codexOpts)
	source := "codex/usage"
	if err != nil {
		whamOpts := base
		whamOpts.URL = api.WhamUsageURL
		raw, err = get(ctx, whamOpts)
		if err != nil {
			return Snapshot{}, fmt.Errorf("codex usage: fetch: %w", err)
		}
		source = "wham/usage"
	}
	return NormalizeJSON(raw, source)
}
