package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func writeGrokHome(t *testing.T, access string, cache []byte) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), fixtureAuthJSON(access), 0o600); err != nil {
		t.Fatal(err)
	}
	if cache == nil {
		cache = []byte(`{"models":{"grok-4.6":{"info":{"id":"grok-4.6","name":"Grok 4.6","api_backend":"responses","context_window":500000,"supported_in_api":true}}}}`)
	}
	if err := os.WriteFile(filepath.Join(dir, "models_cache.json"), cache, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestNewHandler_MissingAuth(t *testing.T) {
	_, err := NewHandler(HandlerOpts{Home: t.TempDir()})
	if err == nil {
		t.Fatal("want auth error")
	}
}

func TestNewHandler_ForwardsResponses(t *testing.T) {
	var gotAuth, gotTokenAuth, gotOverride, gotVersion, gotPath string
	var gotBody map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotTokenAuth = r.Header.Get(HeaderTokenAuth)
		gotOverride = r.Header.Get(HeaderModelOverride)
		gotVersion = r.Header.Get(HeaderClientVersion)
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer upstream.Close()

	home := writeGrokHome(t, "session-token", nil)
	h, err := NewHandler(HandlerOpts{
		Home:    home,
		BaseURL: upstream.URL,
		HTTP:    upstream.Client(),
		Ensure: func(ctx context.Context, opts EnsureOpts) (Auth, error) {
			return opts.Auth, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/responses", "application/json", strings.NewReader(`{"model":"grok-4.6:high","input":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if gotPath != "/v1/responses" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer session-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotTokenAuth != TokenAuthValue {
		t.Fatalf("token auth = %q", gotTokenAuth)
	}
	if gotOverride != "grok-4.6" {
		t.Fatalf("override = %q", gotOverride)
	}
	if gotVersion != DefaultClientVersion {
		t.Fatalf("client version = %q", gotVersion)
	}
	if gotBody["model"] != "grok-4.6" {
		t.Fatalf("body model = %v", gotBody["model"])
	}
	if gotBody["stream"] != true {
		t.Fatalf("stream = %v", gotBody["stream"])
	}
	reasoning, _ := gotBody["reasoning"].(map[string]any)
	if reasoning["effort"] != "high" {
		t.Fatalf("reasoning = %v", gotBody["reasoning"])
	}
}

func TestNewHandler_RetriesUnauthorized(t *testing.T) {
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			if r.Header.Get("Authorization") != "Bearer old" {
				t.Errorf("first auth = %q", r.Header.Get("Authorization"))
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Authorization") != "Bearer new" {
			t.Errorf("retry auth = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	home := writeGrokHome(t, "old", nil)
	h, err := NewHandler(HandlerOpts{
		Home:    home,
		BaseURL: upstream.URL,
		HTTP:    upstream.Client(),
		Ensure: func(ctx context.Context, opts EnsureOpts) (Auth, error) {
			if opts.ForceRefresh {
				return Auth{AccessToken: "new"}, nil
			}
			return opts.Auth, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/responses", "application/json", strings.NewReader(`{"model":"grok-4.6"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestNewHandler_ModelsV2(t *testing.T) {
	home := writeGrokHome(t, "t", nil)
	h, err := NewHandler(HandlerOpts{
		Home:     home,
		Endpoint: "http://localhost:8893/v1",
		Ensure: func(ctx context.Context, opts EnsureOpts) (Auth, error) {
			return opts.Auth, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/models-v2")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var body ModelsV2Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 1 || body.Data[0].ID != "grok-4.6" || body.Data[0].APIBackend != "responses" {
		t.Fatalf("%+v", body)
	}
}

func TestNewHandler_ModelsMissingCache(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), fixtureAuthJSON("t"), 0o600); err != nil {
		t.Fatal(err)
	}
	h, err := NewHandler(HandlerOpts{Home: dir})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status %d", resp.StatusCode)
	}
}
