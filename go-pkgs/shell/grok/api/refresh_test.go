package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRefreshAccessToken_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Fatalf("grant_type = %q", r.Form.Get("grant_type"))
		}
		if r.Form.Get("refresh_token") != "refresh-token-fixture" {
			t.Fatalf("refresh_token unexpected")
		}
		if r.Form.Get("client_id") != "client-fixture" {
			t.Fatalf("client_id = %q", r.Form.Get("client_id"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "rotated-refresh",
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	fixed := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	got, err := RefreshAccessToken(context.Background(), RefreshOpts{
		Auth: Auth{
			AccessToken:  "old",
			RefreshToken: "refresh-token-fixture",
			OIDCIssuer:   "https://auth.x.ai",
			OIDCClientID: "client-fixture",
			AuthKey:      "https://auth.x.ai::client-fixture",
		},
		HTTPClient: srv.Client(),
		TokenURL:   srv.URL,
		Now:        func() time.Time { return fixed },
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "new-access-token" || got.RefreshToken != "rotated-refresh" {
		t.Fatalf("got %+v", got)
	}
	if got.ExpiresAt.Unix() != fixed.Add(time.Hour).Unix() {
		t.Fatalf("ExpiresAt = %v", got.ExpiresAt)
	}
}

func TestRefreshAccessToken_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	_, err := RefreshAccessToken(context.Background(), RefreshOpts{
		Auth: Auth{
			RefreshToken: "refresh-token-fixture",
			OIDCClientID: "client-fixture",
			OIDCIssuer:   "https://auth.x.ai",
		},
		HTTPClient: srv.Client(),
		TokenURL:   srv.URL,
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), "invalid_grant") {
		t.Fatal("error must not include response body")
	}
}

func TestEnsureAccessToken_ForceRefresh(t *testing.T) {
	var refreshed bool
	auth := Auth{
		AccessToken:  "still-valid",
		RefreshToken: "refresh-token-fixture",
		OIDCClientID: "client-fixture",
		OIDCIssuer:   "https://auth.x.ai",
		AuthKey:      "k",
		ExpiresAt:    time.Now().Add(2 * time.Hour),
	}
	got, err := EnsureAccessToken(context.Background(), EnsureOpts{
		Auth:         auth,
		ForceRefresh: true,
		Refresh: func(ctx context.Context, opts RefreshOpts) (Auth, error) {
			refreshed = true
			out := opts.Auth
			out.AccessToken = "forced-new"
			out.ExpiresAt = time.Now().Add(time.Hour)
			return out, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !refreshed || got.AccessToken != "forced-new" {
		t.Fatalf("refreshed=%v got=%+v", refreshed, got)
	}
}

func TestEnsureAccessToken_ExpiredRefreshesAndSaves(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, fixtureAuthJSON("old-access"), 0o600); err != nil {
		t.Fatal(err)
	}
	auth, err := LoadAuth(path)
	if err != nil {
		t.Fatal(err)
	}
	auth.ExpiresAt = time.Now().Add(-time.Minute)

	got, err := EnsureAccessToken(context.Background(), EnsureOpts{
		Auth:     auth,
		AuthPath: path,
		Refresh: func(ctx context.Context, opts RefreshOpts) (Auth, error) {
			out := opts.Auth
			out.AccessToken = "refreshed-access"
			out.RefreshToken = "refreshed-refresh"
			out.ExpiresAt = time.Now().Add(time.Hour)
			return out, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "refreshed-access" {
		t.Fatalf("got %+v", got)
	}
	loaded, err := LoadAuth(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AccessToken != "refreshed-access" || loaded.RefreshToken != "refreshed-refresh" {
		t.Fatalf("persisted %+v", loaded)
	}
}

func TestEnsureAccessToken_ValidSkipsRefresh(t *testing.T) {
	auth := Auth{
		AccessToken: "ok",
		ExpiresAt:   time.Now().Add(2 * time.Hour),
	}
	got, err := EnsureAccessToken(context.Background(), EnsureOpts{
		Auth: auth,
		Refresh: func(ctx context.Context, opts RefreshOpts) (Auth, error) {
			t.Fatal("should not refresh")
			return Auth{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "ok" {
		t.Fatalf("got %+v", got)
	}
}
