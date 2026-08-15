package testconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gitops "github.com/xhd2015/gitops/git"
)

const configFileName = ".test-config.json"

// Config holds test configuration keys
type Config struct {
	AIProvider      string `json:"ai_provider"`
	AIAPIKey        string `json:"ai_api_key"`
	AIBaseURL       string `json:"ai_base_url"`
	AIModel         string `json:"ai_model"`
	BochaAPIKey     string `json:"bocha_api_key"`
	SerpAPIKey      string `json:"serpapi_key"`
	GoogleSearchKey string `json:"google_search_key"`
	GoogleSearchCX  string `json:"google_search_cx"`
}

// Load reads .test-config.json from the go-pkgs module root.
// It skips the test (with setup instructions) if the file is missing.
func Load(t *testing.T) *Config {
	t.Helper()

	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("Failed to find module root: %v", err)
	}

	path := filepath.Join(root, configFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf(`Test config file not found: %s

Please create it with your API keys:

  cat > %s << 'EOF'
{
  "ai_provider": "deepseek",
  "ai_api_key": "your-api-key",
  "ai_base_url": "",
  "ai_model": "",
  "bocha_api_key": "",
  "serpapi_key": "",
  "google_search_key": "",
  "google_search_cx": ""
}
EOF

How to get the keys:

  ai_api_key (required):
    - DeepSeek: Sign up at https://platform.deepseek.com, go to API Keys to create one
    - OpenAI: Sign up at https://platform.openai.com, go to API Keys to create one

  Web search (at least one required for web search tests):

  bocha_api_key (recommended for China):
    - Sign up at https://open.bochaai.com, create an API key
    - Best option for users in China (no GFW issues)

  serpapi_key:
    - Sign up at https://serpapi.com, go to Dashboard > API Key
    - Free tier: 100 searches/month
`, path, path)
		}
		t.Fatalf("Failed to read test config: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to parse test config %s: %v", path, err)
	}
	return &cfg
}

// RequireAI loads config and fails the test if AI API key is not configured
func RequireAI(t *testing.T) *Config {
	t.Helper()
	cfg := Load(t)
	if cfg.AIAPIKey == "" {
		t.Fatal("ai_api_key is not set in .test-config.json")
	}
	return cfg
}

// RequireWebSearch loads config and fails if no web search API is configured
func RequireWebSearch(t *testing.T) *Config {
	t.Helper()
	cfg := Load(t)
	if cfg.BochaAPIKey == "" && cfg.SerpAPIKey == "" && cfg.GoogleSearchKey == "" {
		t.Fatal("No web search API configured in .test-config.json (set bocha_api_key, serpapi_key, or google_search_key)")
	}
	return cfg
}

func findModuleRoot() (string, error) {
	// Try git rev-parse first to find the repo root, then look for go-pkgs
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	out, err := gitops.ShowToplevel(cwd)
	if err == nil {
		repoRoot := strings.TrimSpace(out)
		goPkgsRoot := filepath.Join(repoRoot, "dot-pkgs", "go-pkgs")
		if _, err := os.Stat(filepath.Join(goPkgsRoot, "go.mod")); err == nil {
			return goPkgsRoot, nil
		}
	}

	// Fallback: walk up from current directory looking for go.mod with our module name
	dir := cwd
	for {
		modFile := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modFile)
		if err == nil && strings.Contains(string(data), "github.com/xhd2015/dot-pkgs/go-pkgs") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find go-pkgs module root")
}
