// Package api is the low-level ChatGPT/Codex backend transport for Codex CLI
// auth and HTTP helpers. It does not invent product "usage %" semantics.
package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Auth is the ChatGPT-login credential bundle from auth.json.
// Callers must never log AccessToken, RefreshToken, or IDToken.
type Auth struct {
	AuthMode     string
	AccessToken  string
	AccountID    string
	RefreshToken string
	IDToken      string
	LastRefresh  string
}

// DefaultAuthPath returns $CODEX_HOME/auth.json or ~/.codex/auth.json.
func DefaultAuthPath() (string, error) {
	if home := strings.TrimSpace(os.Getenv("CODEX_HOME")); home != "" {
		return filepath.Join(home, "auth.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("codex api: user home: %w", err)
	}
	return filepath.Join(home, ".codex", "auth.json"), nil
}

// LoadAuth reads and parses a Codex auth.json file.
func LoadAuth(path string) (Auth, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Auth{}, fmt.Errorf("codex api: empty auth path")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Auth{}, fmt.Errorf("codex api: read auth: %w", err)
	}
	return ParseAuthJSON(raw)
}

// ParseAuthJSON parses auth.json bytes into Auth.
func ParseAuthJSON(raw []byte) (Auth, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return Auth{}, fmt.Errorf("codex api: empty auth json")
	}
	var body struct {
		AuthMode     string `json:"auth_mode"`
		LastRefresh  string `json:"last_refresh"`
		OPENAIAPIKey any    `json:"OPENAI_API_KEY"`
		Tokens       struct {
			AccessToken  string `json:"access_token"`
			AccountID    string `json:"account_id"`
			RefreshToken string `json:"refresh_token"`
			IDToken      string `json:"id_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return Auth{}, fmt.Errorf("codex api: decode auth json: %w", err)
	}
	auth := Auth{
		AuthMode:     strings.TrimSpace(body.AuthMode),
		AccessToken:  strings.TrimSpace(body.Tokens.AccessToken),
		AccountID:    strings.TrimSpace(body.Tokens.AccountID),
		RefreshToken: strings.TrimSpace(body.Tokens.RefreshToken),
		IDToken:      strings.TrimSpace(body.Tokens.IDToken),
		LastRefresh:  strings.TrimSpace(body.LastRefresh),
	}
	if auth.AccessToken == "" {
		return Auth{}, fmt.Errorf("codex api: auth json missing access_token")
	}
	if auth.AccountID == "" {
		return Auth{}, fmt.Errorf("codex api: auth json missing account_id")
	}
	return auth, nil
}
