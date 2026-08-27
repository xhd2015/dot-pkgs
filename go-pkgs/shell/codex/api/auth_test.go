package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAuthJSON_OK(t *testing.T) {
	raw := []byte(`{
  "auth_mode": "chatgpt",
  "last_refresh": "2026-01-02T03:04:05.000000Z",
  "tokens": {
    "access_token": "access-token-fixture",
    "account_id": "00000000-0000-4000-8000-000000000001",
    "refresh_token": "refresh-token-fixture",
    "id_token": "id-token-fixture"
  }
}`)
	auth, err := ParseAuthJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if auth.AuthMode != "chatgpt" || auth.AccessToken != "access-token-fixture" {
		t.Fatalf("auth = %+v", auth)
	}
	if auth.AccountID != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("account = %q", auth.AccountID)
	}
}

func TestParseAuthJSON_MissingAccessToken(t *testing.T) {
	raw := []byte(`{"auth_mode":"chatgpt","tokens":{"account_id":"00000000-0000-4000-8000-000000000001"}}`)
	_, err := ParseAuthJSON(raw)
	if err == nil || !strings.Contains(err.Error(), "access_token") {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), "access-token-fixture") {
		t.Fatal("error must not include token material")
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
	body := `{
  "auth_mode": "chatgpt",
  "tokens": {
    "access_token": "access-token-fixture",
    "account_id": "00000000-0000-4000-8000-000000000001"
  }
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
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
