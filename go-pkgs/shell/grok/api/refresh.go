package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RefreshOpts configures RefreshAccessToken.
type RefreshOpts struct {
	Auth Auth
	// HTTPClient nil → http.DefaultClient.
	HTTPClient *http.Client
	// TokenURL empty → {OIDCIssuer}/oauth2/token.
	TokenURL string
	// Now nil → time.Now.
	Now func() time.Time
}

// RefreshAccessToken exchanges refresh_token for a new access token.
// It does not persist; callers use SaveAuth. Never log tokens.
func RefreshAccessToken(ctx context.Context, opts RefreshOpts) (Auth, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	auth := opts.Auth
	if strings.TrimSpace(auth.RefreshToken) == "" {
		return Auth{}, fmt.Errorf("grok api: missing refresh_token")
	}
	clientID := strings.TrimSpace(auth.OIDCClientID)
	if clientID == "" {
		return Auth{}, fmt.Errorf("grok api: missing oidc_client_id")
	}
	tokenURL := strings.TrimSpace(opts.TokenURL)
	if tokenURL == "" {
		issuer := strings.TrimSpace(auth.OIDCIssuer)
		if issuer == "" {
			return Auth{}, fmt.Errorf("grok api: missing oidc_issuer")
		}
		tokenURL = strings.TrimRight(issuer, "/") + "/oauth2/token"
	}

	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", auth.RefreshToken)
	form.Set("client_id", clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Auth{}, fmt.Errorf("grok api: refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return Auth{}, fmt.Errorf("grok api: refresh: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Auth{}, fmt.Errorf("grok api: refresh read: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Auth{}, fmt.Errorf("grok api: refresh HTTP %d", resp.StatusCode)
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return Auth{}, fmt.Errorf("grok api: refresh decode: %w", err)
	}
	access := strings.TrimSpace(tok.AccessToken)
	if access == "" {
		return Auth{}, fmt.Errorf("grok api: refresh response missing access_token")
	}
	out := auth
	out.AccessToken = access
	if rt := strings.TrimSpace(tok.RefreshToken); rt != "" {
		out.RefreshToken = rt
	}
	if tok.ExpiresIn > 0 {
		out.ExpiresAt = now().UTC().Add(time.Duration(tok.ExpiresIn) * time.Second)
	} else if t, ok := jwtExpiry(access); ok {
		out.ExpiresAt = t
	}
	return out, nil
}

// EnsureOpts configures EnsureAccessToken.
type EnsureOpts struct {
	Auth Auth
	// AuthPath is used with SaveAuth when a refresh occurs. Empty skips persist
	// unless Save is set.
	AuthPath string
	// ForceRefresh always refreshes even when the access token is still valid.
	ForceRefresh bool
	// Skew empty/zero → DefaultSkew.
	Skew time.Duration
	// Now nil → time.Now.
	Now func() time.Time
	// HTTPClient passed to Refresh when Refresh is nil.
	HTTPClient *http.Client
	// TokenURL passed to Refresh when Refresh is nil.
	TokenURL string
	// Refresh injectable; nil → RefreshAccessToken.
	Refresh func(ctx context.Context, opts RefreshOpts) (Auth, error)
	// Save injectable; nil → SaveAuth when AuthPath is set.
	Save func(path string, auth Auth) error
}

// EnsureAccessToken returns auth with a usable access token.
// Refreshes when ForceRefresh is set or AccessTokenExpired; then persists
// when AuthPath or Save is configured.
func EnsureAccessToken(ctx context.Context, opts EnsureOpts) (Auth, error) {
	auth := opts.Auth
	if strings.TrimSpace(auth.AccessToken) == "" && !opts.ForceRefresh {
		return Auth{}, fmt.Errorf("grok api: missing access token")
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	skew := opts.Skew
	if skew == 0 {
		skew = DefaultSkew
	}

	need := opts.ForceRefresh || AccessTokenExpired(auth, nowFn(), skew)
	if !need {
		return auth, nil
	}
	if strings.TrimSpace(auth.RefreshToken) == "" {
		return Auth{}, fmt.Errorf("grok api: access token expired; missing refresh_token")
	}

	refresh := opts.Refresh
	if refresh == nil {
		refresh = RefreshAccessToken
	}
	updated, err := refresh(ctx, RefreshOpts{
		Auth:       auth,
		HTTPClient: opts.HTTPClient,
		TokenURL:   opts.TokenURL,
		Now:        nowFn,
	})
	if err != nil {
		return Auth{}, err
	}

	save := opts.Save
	path := strings.TrimSpace(opts.AuthPath)
	if save == nil && path != "" {
		save = SaveAuth
	}
	if save != nil && path != "" {
		if err := save(path, updated); err != nil {
			return Auth{}, err
		}
	}
	return updated, nil
}
