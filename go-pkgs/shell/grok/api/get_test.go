package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetJSON_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access-token-fixture" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Fatal("missing User-Agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"config":{"used":{"val":1}}}`))
	}))
	defer srv.Close()

	body, err := GetJSON(context.Background(), GetOpts{
		URL: srv.URL,
		Auth: Auth{
			AccessToken: "access-token-fixture",
		},
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"used"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestGetJSON_HTTP403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("<html>denied</html>"))
	}))
	defer srv.Close()

	_, err := GetJSON(context.Background(), GetOpts{
		URL:        srv.URL,
		Auth:       Auth{AccessToken: "access-token-fixture"},
		HTTPClient: srv.Client(),
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), "<html>") {
		t.Fatal("error must not include response body")
	}
	if !IsUnauthorized(err) {
		t.Fatal("want IsUnauthorized")
	}
}

func TestGetJSON_MissingAuth(t *testing.T) {
	if _, err := GetJSON(context.Background(), GetOpts{URL: "http://example.com"}); err == nil {
		t.Fatal("want missing token error")
	}
}
