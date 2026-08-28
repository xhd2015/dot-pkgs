// Package api is the low-level Grok CLI auth and HTTP transport.
// It does not invent product "usage %" semantics.
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Auth is the OIDC credential bundle from ~/.grok/auth.json.
// Callers must never log AccessToken or RefreshToken.
type Auth struct {
	// AuthKey is the top-level map key (issuer::client_id) for SaveAuth.
	AuthKey  string
	AuthMode string
	// AccessToken is the JWT stored as JSON field "key".
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Email        string
	UserID       string
	OIDCIssuer   string
	OIDCClientID string
}

type authEntry struct {
	Key          string `json:"key"`
	AuthMode     string `json:"auth_mode"`
	CreateTime   string `json:"create_time"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	OIDCIssuer   string `json:"oidc_issuer"`
	OIDCClientID string `json:"oidc_client_id"`
	PrincipalID  string `json:"principal_id"`
}

// DefaultAuthPath returns $GROK_HOME/auth.json or ~/.grok/auth.json.
func DefaultAuthPath() (string, error) {
	if home := strings.TrimSpace(os.Getenv("GROK_HOME")); home != "" {
		return filepath.Join(home, "auth.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("grok api: user home: %w", err)
	}
	return filepath.Join(home, ".grok", "auth.json"), nil
}

// LoadAuth reads and parses a Grok auth.json file.
func LoadAuth(path string) (Auth, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Auth{}, fmt.Errorf("grok api: empty auth path")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Auth{}, fmt.Errorf("grok api: read auth: %w", err)
	}
	return ParseAuthJSON(raw)
}

// ParseAuthJSON parses auth.json bytes into Auth.
// The file is a map keyed by "issuer::client_id"; the first OIDC entry with a
// non-empty key wins (keys sorted for stability when multiple exist).
func ParseAuthJSON(raw []byte) (Auth, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return Auth{}, fmt.Errorf("grok api: empty auth json")
	}
	var body map[string]authEntry
	if err := json.Unmarshal(raw, &body); err != nil {
		return Auth{}, fmt.Errorf("grok api: decode auth json: %w", err)
	}
	if len(body) == 0 {
		return Auth{}, fmt.Errorf("grok api: auth json has no entries")
	}
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var chosenKey string
	var chosen authEntry
	for _, k := range keys {
		e := body[k]
		if strings.TrimSpace(e.Key) == "" {
			continue
		}
		mode := strings.ToLower(strings.TrimSpace(e.AuthMode))
		if mode == "" || mode == "oidc" {
			chosenKey, chosen = k, e
			break
		}
		if chosenKey == "" {
			chosenKey, chosen = k, e
		}
	}
	if chosenKey == "" || strings.TrimSpace(chosen.Key) == "" {
		return Auth{}, fmt.Errorf("grok api: auth json missing access token (key)")
	}

	auth := Auth{
		AuthKey:      chosenKey,
		AuthMode:     strings.TrimSpace(chosen.AuthMode),
		AccessToken:  strings.TrimSpace(chosen.Key),
		RefreshToken: strings.TrimSpace(chosen.RefreshToken),
		Email:        strings.TrimSpace(chosen.Email),
		UserID:       strings.TrimSpace(chosen.UserID),
		OIDCIssuer:   strings.TrimSpace(chosen.OIDCIssuer),
		OIDCClientID: strings.TrimSpace(chosen.OIDCClientID),
	}
	if auth.UserID == "" {
		auth.UserID = strings.TrimSpace(chosen.PrincipalID)
	}
	if auth.OIDCIssuer == "" || auth.OIDCClientID == "" {
		if iss, cid, ok := splitAuthKey(chosenKey); ok {
			if auth.OIDCIssuer == "" {
				auth.OIDCIssuer = iss
			}
			if auth.OIDCClientID == "" {
				auth.OIDCClientID = cid
			}
		}
	}
	if t, ok := parseExpiresAt(chosen.ExpiresAt); ok {
		auth.ExpiresAt = t
	} else if t, ok := jwtExpiry(auth.AccessToken); ok {
		auth.ExpiresAt = t
	}
	return auth, nil
}

func splitAuthKey(key string) (issuer, clientID string, ok bool) {
	// Expected: https://auth.x.ai::b1a00492-...
	i := strings.Index(key, "::")
	if i <= 0 || i+2 >= len(key) {
		return "", "", false
	}
	return key[:i], key[i+2:], true
}

func parseExpiresAt(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// jwtExpiry decodes exp from an unverified JWT payload.
func jwtExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Some encoders include padding.
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return time.Time{}, false
		}
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp <= 0 {
		return time.Time{}, false
	}
	return time.Unix(claims.Exp, 0).UTC(), true
}

// DefaultSkew is how early AccessTokenExpired treats a token as expired.
const DefaultSkew = 60 * time.Second

// AccessTokenExpired reports whether auth needs a refresh at now.
// If ExpiresAt is zero, returns false (caller may still ForceRefresh).
func AccessTokenExpired(auth Auth, now time.Time, skew time.Duration) bool {
	if auth.ExpiresAt.IsZero() {
		return false
	}
	if skew < 0 {
		skew = 0
	}
	if now.IsZero() {
		now = time.Now()
	}
	return !now.Before(auth.ExpiresAt.Add(-skew))
}

// SaveAuth updates the access/refresh tokens for auth.AuthKey in path.
// Other entry fields are preserved. Writes atomically with mode 0600.
func SaveAuth(path string, auth Auth) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("grok api: empty auth path")
	}
	if strings.TrimSpace(auth.AuthKey) == "" {
		return fmt.Errorf("grok api: empty auth key")
	}
	if strings.TrimSpace(auth.AccessToken) == "" {
		return fmt.Errorf("grok api: missing access token (key)")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("grok api: read auth for save: %w", err)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return fmt.Errorf("grok api: decode auth for save: %w", err)
	}
	entryRaw, ok := body[auth.AuthKey]
	if !ok {
		return fmt.Errorf("grok api: auth key not found for save")
	}
	var entry map[string]any
	if err := json.Unmarshal(entryRaw, &entry); err != nil {
		return fmt.Errorf("grok api: decode auth entry for save: %w", err)
	}
	entry["key"] = auth.AccessToken
	if auth.RefreshToken != "" {
		entry["refresh_token"] = auth.RefreshToken
	}
	if !auth.ExpiresAt.IsZero() {
		entry["expires_at"] = auth.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}
	updated, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("grok api: encode auth entry: %w", err)
	}
	body[auth.AuthKey] = updated

	out, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return fmt.Errorf("grok api: encode auth json: %w", err)
	}
	out = append(out, '\n')
	return writeFileAtomic(path, out, 0o600)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".auth-*.tmp")
	if err != nil {
		return fmt.Errorf("grok api: create temp auth: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("grok api: write temp auth: %w", err)
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("grok api: chmod temp auth: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("grok api: close temp auth: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("grok api: rename auth: %w", err)
	}
	return nil
}
