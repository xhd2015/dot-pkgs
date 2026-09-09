package qemu

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds qemu guest and guest-cloudflared settings.
type Config struct {
	Dir         string // guest host dir
	User        string // default debian
	SSHPort     int    // default 22221
	MemMB       int
	CPUs        int
	BackingName string
	BackingURL  string
	PkgVersion  string // debian qemu pin (optional)
	CFOriginURL string
	CFBin       string
	CFCert      string
	CFLog       string
	CFPid       string
	CFVersion   string
	Accel       string // "tcg" or "hvf" etc display label
}

// FileConfig is ~/.ai-critic/qemu.json.
type FileConfig struct {
	Enabled bool `json:"enabled"`
}

// WorkDevConfig returns the work.dev CodeLens admin qemu defaults.
func WorkDevConfig() Config {
	return Config{
		Dir:         "/root/qemu-guest",
		User:        "debian",
		SSHPort:     22221,
		MemMB:       1024,
		CPUs:        2,
		BackingName: "debian-12-genericcloud-amd64.qcow2",
		BackingURL:  "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2",
		PkgVersion:  "1:5.2+dfsg-11+deb11u3",
		CFOriginURL: "http://127.0.0.1:8080",
		CFBin:       "/usr/local/bin/cloudflared",
		CFCert:      "/root/.cloudflared/cert.pem",
		CFLog:       "/var/log/work-qemu-cloudflared.log",
		CFPid:       "/var/run/work-qemu-cloudflared.pid",
		CFVersion:   "2025.11.1",
		Accel:       "tcg",
	}
}

// AiCriticConfig returns local-oriented defaults under ~/.ai-critic/qemu.
// Accel is "auto" (prefer hvf, else tcg). CF origin defaults to the host
// SLIRP gateway path for a server on the host.
func AiCriticConfig() Config {
	c := WorkDevConfig()
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	c.Dir = filepath.Join(home, ".ai-critic", "qemu")
	c.Accel = "auto"
	c.CFOriginURL = "http://10.0.2.2:23712"
	return c
}

// DefaultFilePath returns ~/.ai-critic/qemu.json for the given home.
func DefaultFilePath(home string) string {
	return filepath.Join(home, ".ai-critic", "qemu.json")
}

// LoadFileConfig reads a FileConfig JSON file.
func LoadFileConfig(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, err
	}
	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, err
	}
	return cfg, nil
}

// SaveFileConfig writes a FileConfig JSON file, creating parent dirs as needed.
func SaveFileConfig(path string, cfg FileConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// IsEnabled reports whether path enables qemu. Missing file or read/parse
// errors yield false.
func IsEnabled(path string) bool {
	cfg, err := LoadFileConfig(path)
	if err != nil {
		return false
	}
	return cfg.Enabled
}
