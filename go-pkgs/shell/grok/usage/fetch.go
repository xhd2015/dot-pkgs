// Package usage fetches and normalizes Grok account billing via shell/grok/api.
package usage

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/api"
)

// FetchOpts configures Fetch.
type FetchOpts struct {
	// AuthPath empty → api.DefaultAuthPath(). Ignored when Auth has AccessToken
	// and ForceRefresh/expiry handling still may need AuthPath to persist.
	AuthPath string
	// Auth when non-zero AccessToken skips file load (tests).
	Auth api.Auth
	// ForceRefresh always refreshes the access token before GET.
	ForceRefresh bool
	// Skew for expiry check; 0 → api.DefaultSkew.
	Skew time.Duration
	// HTTPClient nil → http.DefaultClient (via api).
	HTTPClient *http.Client
	// UserAgent empty → api.DefaultUserAgent.
	UserAgent string
	// LoadAuth injectable; nil → api.LoadAuth.
	LoadAuth func(path string) (api.Auth, error)
	// Ensure injectable; nil → api.EnsureAccessToken.
	Ensure func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error)
	// GetJSON injectable; nil → api.GetJSON. When set, URL is still chosen by Fetch.
	GetJSON func(ctx context.Context, opts api.GetOpts) ([]byte, error)
	// Now injectable for expiry; nil → time.Now.
	Now func() time.Time
}

// Fetch loads auth, ensures a fresh access token, GETs billing, then normalizes.
// On HTTP 401/403, refreshes once (forced) and retries the GET once.
func Fetch(ctx context.Context, opts FetchOpts) (Snapshot, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	authPath := strings.TrimSpace(opts.AuthPath)
	auth := opts.Auth
	if strings.TrimSpace(auth.AccessToken) == "" {
		if authPath == "" {
			var err error
			authPath, err = api.DefaultAuthPath()
			if err != nil {
				return Snapshot{}, err
			}
		}
		load := opts.LoadAuth
		if load == nil {
			load = api.LoadAuth
		}
		var err error
		auth, err = load(authPath)
		if err != nil {
			return Snapshot{}, err
		}
	}

	ensure := opts.Ensure
	if ensure == nil {
		ensure = api.EnsureAccessToken
	}
	ensureOpts := api.EnsureOpts{
		Auth:         auth,
		AuthPath:     authPath,
		ForceRefresh: opts.ForceRefresh,
		Skew:         opts.Skew,
		HTTPClient:   opts.HTTPClient,
		Now:          opts.Now,
	}
	auth, err := ensure(ctx, ensureOpts)
	if err != nil {
		return Snapshot{}, err
	}

	get := opts.GetJSON
	if get == nil {
		get = api.GetJSON
	}
	base := api.GetOpts{
		URL:        api.BillingURL,
		Auth:       auth,
		HTTPClient: opts.HTTPClient,
		UserAgent:  opts.UserAgent,
	}

	raw, err := get(ctx, base)
	if err != nil && api.IsUnauthorized(err) {
		// One forced refresh + retry.
		auth, err2 := ensure(ctx, api.EnsureOpts{
			Auth:         auth,
			AuthPath:     authPath,
			ForceRefresh: true,
			Skew:         opts.Skew,
			HTTPClient:   opts.HTTPClient,
			Now:          opts.Now,
		})
		if err2 != nil {
			return Snapshot{}, fmt.Errorf("grok usage: fetch: %w", err)
		}
		base.Auth = auth
		raw, err = get(ctx, base)
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("grok usage: fetch: %w", err)
	}

	snap, err := NormalizeJSON(raw, "billing")
	if err != nil {
		return Snapshot{}, err
	}
	if snap.Email == "" {
		snap.Email = auth.Email
	}
	return snap, nil
}
