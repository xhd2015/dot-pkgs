package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureCache() ModelsCache {
	return ModelsCache{
		ETag: `"etag-1"`,
		Models: map[string]CachedModel{
			"grok-4.6": {Info: ModelInfo{
				ID:             "grok-4.6",
				Name:           "Grok 4.6",
				Description:    "latest",
				APIBackend:     "responses",
				ContextWindow:  500000,
				SupportedInAPI: true,
				ReasoningEfforts: []ReasoningEffort{
					{Value: "xhigh"},
					{Value: "high", Default: true},
					{Value: "medium"},
					{Value: "low"},
				},
			}},
			"hidden": {Info: ModelInfo{ID: "hidden", Hidden: true, SupportedInAPI: true}},
			"off":    {Info: ModelInfo{ID: "off", SupportedInAPI: false}},
		},
	}
}

func TestModelsV2(t *testing.T) {
	resp := ModelsV2(fixtureCache(), "http://localhost:8893/v1")
	if len(resp.Data) != 1 {
		t.Fatalf("len = %d", len(resp.Data))
	}
	m := resp.Data[0]
	if m.ID != "grok-4.6" || m.APIBackend != "responses" || m.ContextWindow != 500000 {
		t.Fatalf("entry = %+v", m)
	}
	if m.BaseURL != "http://localhost:8893/v1" {
		t.Fatalf("base = %q", m.BaseURL)
	}
}

func TestOpenAIModels(t *testing.T) {
	resp := OpenAIModels(fixtureCache())
	if resp.Object != "list" || len(resp.Data) != 1 || resp.Data[0].ID != "grok-4.6" {
		t.Fatalf("%+v", resp)
	}
}

func TestCodexConfig(t *testing.T) {
	out := CodexConfig("http://localhost:8893/v1", fixtureCache())
	for _, want := range []string{
		`model = "grok-4.6"`,
		`model_provider = "llm-proxy"`,
		`base_url = "http://localhost:8893/v1"`,
		`wire_api = "responses"`,
		"#   grok-4.6:high",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestDSHSettingsYAML(t *testing.T) {
	out := DSHSettingsYAML("http://localhost:8893/v1", fixtureCache())
	for _, want := range []string{
		"api: openai-responses",
		"baseURL: http://127.0.0.1:8893/v1",
		"id: grok-4.6",
		"contextWindow: 500000",
		"xhigh: xhigh",
		"apiKeyEnv: GROK_API_KEY",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "baseURL: http://localhost") {
		t.Errorf("DSH baseURL should use 127.0.0.1:\n%s", out)
	}
}

func TestDSHBaseURL(t *testing.T) {
	got := DSHBaseURL("http://localhost:8893/v1")
	if got != "http://127.0.0.1:8893/v1" {
		t.Fatalf("got %q", got)
	}
}

func TestReadModelsCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models_cache.json")
	raw := []byte(`{"etag":"e","models":{"grok-4.6":{"info":{"id":"grok-4.6","supported_in_api":true,"context_window":1}}}}`)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	cache, err := ReadModelsCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if cache.ETag != "e" || cache.Models["grok-4.6"].Info.ContextWindow != 1 {
		t.Fatalf("%+v", cache)
	}
}
