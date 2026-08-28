package api

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixtureAuthJSON(access string) []byte {
	body := map[string]any{
		"https://auth.x.ai::00000000-0000-4000-8000-000000000099": map[string]any{
			"key":            access,
			"auth_mode":      "oidc",
			"user_id":        "00000000-0000-4000-8000-000000000001",
			"email":          "user@example.com",
			"refresh_token":  "refresh-token-fixture",
			"expires_at":     "2026-08-28T05:57:49.023323Z",
			"oidc_issuer":    "https://auth.x.ai",
			"oidc_client_id": "00000000-0000-4000-8000-000000000099",
		},
	}
	raw, _ := json.Marshal(body)
	return raw
}

func TestParseAuthJSON_OK(t *testing.T) {
	auth, err := ParseAuthJSON(fixtureAuthJSON("access-token-fixture"))
	if err != nil {
		t.Fatal(err)
	}
	if auth.AccessToken != "access-token-fixture" {
		t.Fatalf("AccessToken = %q", auth.AccessToken)
	}
	if auth.Email != "user@example.com" || auth.RefreshToken != "refresh-token-fixture" {
		t.Fatalf("auth = %+v", auth)
	}
	if auth.OIDCIssuer != "https://auth.x.ai" || auth.OIDCClientID == "" {
		t.Fatalf("oidc = %+v", auth)
	}
	if auth.ExpiresAt.IsZero() {
		t.Fatal("want ExpiresAt")
	}
}

func TestParseAuthJSON_MissingKey(t *testing.T) {
	raw := []byte(`{"https://auth.x.ai::abc":{"auth_mode":"oidc","refresh_token":"r"}}`)
	_, err := ParseAuthJSON(raw)
	if err == nil || !strings.Contains(err.Error(), "key") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseAuthJSON_Empty(t *testing.T) {
	if _, err := ParseAuthJSON(nil); err == nil {
		t.Fatal("want empty error")
	}
	if _, err := ParseAuthJSON([]byte("not-json")); err == nil {
		t.Fatal("want decode error")
	}
}

func TestLoadAuth_FromTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, fixtureAuthJSON("access-token-fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	auth, err := LoadAuth(path)
	if err != nil {
		t.Fatal(err)
	}
	if auth.AccessToken != "access-token-fixture" {
		t.Fatalf("got %+v", auth)
	}
}

func TestLoadAuth_EmptyPath(t *testing.T) {
	if _, err := LoadAuth(""); err == nil {
		t.Fatal("want empty path error")
	}
}

func TestAccessTokenExpired(t *testing.T) {
	now := time.Date(2026, 8, 28, 6, 0, 0, 0, time.UTC)
	auth := Auth{ExpiresAt: now.Add(-time.Minute)}
	if !AccessTokenExpired(auth, now, DefaultSkew) {
		t.Fatal("want expired")
	}
	auth.ExpiresAt = now.Add(2 * time.Hour)
	if AccessTokenExpired(auth, now, DefaultSkew) {
		t.Fatal("want not expired")
	}
	// within skew
	auth.ExpiresAt = now.Add(30 * time.Second)
	if !AccessTokenExpired(auth, now, DefaultSkew) {
		t.Fatal("want expired within skew")
	}
	if AccessTokenExpired(Auth{}, now, DefaultSkew) {
		t.Fatal("zero ExpiresAt → not expired")
	}
}

func TestSaveAuth_UpdatesKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, fixtureAuthJSON("old-access"), 0o600); err != nil {
		t.Fatal(err)
	}
	auth, err := LoadAuth(path)
	if err != nil {
		t.Fatal(err)
	}
	auth.AccessToken = "new-access"
	auth.RefreshToken = "new-refresh"
	auth.ExpiresAt = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if err := SaveAuth(path, auth); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAuth(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "new-access" || got.RefreshToken != "new-refresh" {
		t.Fatalf("got %+v", got)
	}
	if got.ExpiresAt.Unix() != auth.ExpiresAt.Unix() {
		t.Fatalf("ExpiresAt got %v want %v", got.ExpiresAt, auth.ExpiresAt)
	}
}

func TestJWTExpiry(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":1787896668}`))
	tok := "hdr." + payload + ".sig"
	tm, ok := jwtExpiry(tok)
	if !ok || tm.Unix() != 1787896668 {
		t.Fatalf("got %v ok=%v", tm, ok)
	}
}
