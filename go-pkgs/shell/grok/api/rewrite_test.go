package api

import (
	"encoding/json"
	"testing"
)

func TestSplitModelEffort(t *testing.T) {
	tests := []struct {
		in, base, effort string
	}{
		{"grok-4.6", "grok-4.6", ""},
		{"grok-4.6:high", "grok-4.6", "high"},
		{"grok-4.6:xhigh", "grok-4.6", "xhigh"},
		{"grok-4.6:not-an-effort", "grok-4.6:not-an-effort", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		base, effort := SplitModelEffort(tt.in)
		if base != tt.base || effort != tt.effort {
			t.Errorf("SplitModelEffort(%q) = %q, %q; want %q, %q", tt.in, base, effort, tt.base, tt.effort)
		}
	}
}

func TestRewriteRequestJSON_EffortAndStream(t *testing.T) {
	body := []byte(`{"model":"grok-4.6:high","input":[{"role":"user","content":"hi"}]}`)
	model, out, err := RewriteRequestJSON(body)
	if err != nil {
		t.Fatal(err)
	}
	if model != "grok-4.6" {
		t.Fatalf("model = %q", model)
	}
	var data map[string]any
	if err := json.Unmarshal(out, &data); err != nil {
		t.Fatal(err)
	}
	if data["model"] != "grok-4.6" {
		t.Fatalf("body model = %v", data["model"])
	}
	if data["stream"] != true {
		t.Fatalf("stream = %v", data["stream"])
	}
	reasoning, _ := data["reasoning"].(map[string]any)
	if reasoning["effort"] != "high" {
		t.Fatalf("reasoning = %v", data["reasoning"])
	}
}

func TestRewriteRequestJSON_KeepsExistingEffort(t *testing.T) {
	body := []byte(`{"model":"grok-4.6:low","stream":true,"reasoning":{"effort":"high"}}`)
	_, out, err := RewriteRequestJSON(body)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := json.Unmarshal(out, &data); err != nil {
		t.Fatal(err)
	}
	reasoning, _ := data["reasoning"].(map[string]any)
	if reasoning["effort"] != "high" {
		t.Fatalf("effort overwritten: %v", reasoning)
	}
}

func TestRewriteRequestJSON_InvalidJSON(t *testing.T) {
	body := []byte("not-json")
	_, out, err := RewriteRequestJSON(body)
	if err == nil {
		t.Fatal("want error")
	}
	if string(out) != string(body) {
		t.Fatal("want original body")
	}
}
