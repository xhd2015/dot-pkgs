package api

import (
	"net/http"
	"testing"
)

func TestApplySessionHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer client-token")
	ApplySessionHeaders(h, Auth{AccessToken: "session-token"}, "grok-4.6", "1.0.34")
	if got := h.Get("Authorization"); got != "Bearer session-token" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := h.Get(HeaderTokenAuth); got != TokenAuthValue {
		t.Fatalf("token auth = %q", got)
	}
	if got := h.Get(HeaderModelOverride); got != "grok-4.6" {
		t.Fatalf("model override = %q", got)
	}
	if got := h.Get(HeaderClientVersion); got != "1.0.34" {
		t.Fatalf("client version = %q", got)
	}
	if h.Get("User-Agent") == "" {
		t.Fatal("missing User-Agent")
	}
}

func TestApplySessionHeaders_ClearsOverrideWhenModelEmpty(t *testing.T) {
	h := http.Header{}
	h.Set(HeaderModelOverride, "old")
	ApplySessionHeaders(h, Auth{AccessToken: "t"}, "", "")
	if got := h.Get(HeaderModelOverride); got != "" {
		t.Fatalf("override = %q", got)
	}
}
