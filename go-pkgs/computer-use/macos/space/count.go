package space

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MacOSCountMode selects how CountDesktops obtains the Desktop count on macOS.
// It does not open Mission Control (no Spaces Bar UI).
type MacOSCountMode int

const (
	// MacOSCountModePlist reads ~/Library/Preferences/com.apple.spaces.plist
	// (via plutil JSON). Default: no private frameworks.
	MacOSCountModePlist MacOSCountMode = iota
	// MacOSCountModePrivateAPI uses SkyLight CGSCopyManagedDisplaySpaces
	// (heavier; live WindowServer). Darwin only.
	MacOSCountModePrivateAPI
)

// CountOption configures CountDesktops.
type CountOption func(*countConfig)

type countConfig struct {
	mode      MacOSCountMode
	plistPath string // test override; empty = default user prefs path
	// hooks for tests
	readPlistJSON func(path string) ([]byte, error)
	privateCount  func() (int, error)
}

// WithMacOSCountMode selects the macOS counting strategy.
// Default is MacOSCountModePlist.
func WithMacOSCountMode(mode MacOSCountMode) CountOption {
	return func(c *countConfig) {
		c.mode = mode
	}
}

// withPlistPath is for tests only (same package).
func withPlistPath(path string) CountOption {
	return func(c *countConfig) {
		c.plistPath = path
	}
}

// CountDesktops returns the number of normal Mission Control Desktops on the
// primary/main display without opening Mission Control.
//
// Default mode is MacOSCountModePlist. Use WithMacOSCountMode to select the
// private SkyLight API. This helper is optional for callers; Create does not
// call it automatically.
//
// Non-macOS returns ErrUnsupportedPlatform.
func CountDesktops(opts ...CountOption) (int, error) {
	if err := requireDarwin(); err != nil {
		return 0, err
	}
	cfg := &countConfig{mode: MacOSCountModePlist}
	for _, o := range opts {
		if o != nil {
			o(cfg)
		}
	}
	switch cfg.mode {
	case MacOSCountModePrivateAPI:
		if cfg.privateCount != nil {
			return cfg.privateCount()
		}
		return countDesktopsPrivateAPI()
	case MacOSCountModePlist:
		fallthrough
	default:
		return countDesktopsPlist(cfg)
	}
}

func countDesktopsPlist(cfg *countConfig) (int, error) {
	path := cfg.plistPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return 0, fmt.Errorf("space: home dir: %w", err)
		}
		path = filepath.Join(home, "Library", "Preferences", "com.apple.spaces.plist")
	}
	read := cfg.readPlistJSON
	if read == nil {
		read = plutilJSON
	}
	raw, err := read(path)
	if err != nil {
		return 0, fmt.Errorf("space: read spaces plist: %w", err)
	}
	n, err := parseSpacesPlistJSON(raw)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func plutilJSON(path string) ([]byte, error) {
	// Binary plists are not encoding/json-friendly; plutil is always on macOS.
	cmd := exec.Command("plutil", "-convert", "json", "-o", "-", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return stdout.Bytes(), nil
}

// parseSpacesPlistJSON counts type==0 (or missing type) Desktops on the first
// monitor that has a non-empty Spaces list; if none, the first monitor.
func parseSpacesPlistJSON(raw []byte) (int, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return 0, fmt.Errorf("space: parse spaces plist JSON: %w", err)
	}
	sdc, _ := root["SpacesDisplayConfiguration"].(map[string]interface{})
	if sdc == nil {
		return 0, fmt.Errorf("space: spaces plist: missing SpacesDisplayConfiguration")
	}
	md, _ := sdc["Management Data"].(map[string]interface{})
	if md == nil {
		return 0, fmt.Errorf("space: spaces plist: missing Management Data")
	}
	monitors, _ := md["Monitors"].([]interface{})
	if len(monitors) == 0 {
		return 0, fmt.Errorf("space: spaces plist: no monitors")
	}

	// Prefer first monitor with Spaces; else first entry (matches Main in practice).
	var chosen map[string]interface{}
	for _, m := range monitors {
		mm, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		spaces, _ := mm["Spaces"].([]interface{})
		if len(spaces) > 0 {
			chosen = mm
			break
		}
		if chosen == nil {
			chosen = mm
		}
	}
	if chosen == nil {
		return 0, fmt.Errorf("space: spaces plist: no monitor entries")
	}
	spaces, _ := chosen["Spaces"].([]interface{})
	n := 0
	for _, s := range spaces {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		// type 0 = normal Desktop; omit or non-number treated as 0.
		switch t := sm["type"].(type) {
		case nil:
			n++
		case float64:
			if t == 0 {
				n++
			}
		case json.Number:
			if v, err := t.Float64(); err == nil && v == 0 {
				n++
			}
		case int:
			if t == 0 {
				n++
			}
		default:
			// unknown type field — count conservatively as desktop
			n++
		}
	}
	return n, nil
}
