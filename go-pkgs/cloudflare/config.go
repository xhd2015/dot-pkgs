package cloudflare

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// IsUUID checks if a string looks like a UUID (8-4-4-4-12 hex format).
func IsUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// DefaultConfigDir returns ~/.cloudflared.
func DefaultConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	return filepath.Join(homeDir, ".cloudflared"), nil
}

// WriteConfig writes a cloudflared config YAML file, creating parent dirs as needed.
func WriteConfig(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	cfgDir := filepath.Dir(path)
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %v", cfgDir, err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}
	return nil
}

var (
	idFromCreateRe   = regexp.MustCompile(`with id ([a-fA-F0-9-]+)`)
	credFromCreateRe = regexp.MustCompile(`credentials written to (.+\.json)`)
)

func parseCreateOutput(out string) (tunnelID, credFile string) {
	if m := idFromCreateRe.FindStringSubmatch(out); len(m) > 1 {
		tunnelID = m[1]
	}
	if m := credFromCreateRe.FindStringSubmatch(out); len(m) > 1 {
		credFile = m[1]
	}
	return tunnelID, credFile
}

func parseUUIDFromInfo(info string) string {
	for _, line := range splitLines(info) {
		for _, part := range fields(line) {
			if IsUUID(part) {
				return part
			}
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func fields(s string) []string {
	var out []string
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
			i++
		}
		if i >= len(s) {
			break
		}
		j := i
		for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\r' {
			j++
		}
		out = append(out, s[i:j])
		i = j
	}
	return out
}
