// Package space automates macOS Mission Control Desktops (Spaces) via
// Accessibility / AppleScript. It does not parse CLI flags or run follow-up
// commands — callers (e.g. kool macos space) own that.
//
// Example:
//
//	if err := space.Create(nil); err != nil {
//	    log.Fatal(err)
//	}
//	n, err := space.Highest(nil)
package space

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Desktop is a Mission Control Desktop (Space).
type Desktop struct {
	// Number is the 1-based Desktop index shown as "Desktop N".
	Number int
	// Name is typically "Desktop N".
	Name string
}

var (
	// ErrUnsupportedPlatform is returned when not running on macOS.
	ErrUnsupportedPlatform = errors.New("space: unsupported platform (macOS only)")
)

const (
	defaultSettle = 500 * time.Millisecond
	envGOOS       = "DOT_PKGS_SPACE_GOOS"
)

var testGOOS string

// SetGOOSForTest overrides platform detection. Pass "" to reset.
func SetGOOSForTest(goos string) {
	testGOOS = goos
}

func effectiveGOOS() string {
	if testGOOS != "" {
		return testGOOS
	}
	if v := os.Getenv(envGOOS); v != "" {
		return v
	}
	return runtime.GOOS
}

// Config customizes Space operations for tests or alternate runners.
type Config struct {
	// Osascript runs AppleScript. When nil, default osascript is used.
	// args are passed as argv to scripts that use "on run argv" (e.g. Switch).
	Osascript func(script string, args ...string) (stdout string, err error)

	// Settle is the delay after Switch and CreateAndActivate.
	// Zero uses default (500ms). Negative means no sleep.
	Settle time.Duration
}

// Backend is the Space operations surface (real AX or MockBackend).
type Backend interface {
	Create() error
	Switch(n int) error
	List() ([]Desktop, error)
	Highest() (int, error)
}

// Create adds one Desktop via Mission Control UI automation.
func Create(cfg *Config) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	return newAX(cfg).Create()
}

// Switch activates Desktop number (1-based).
func Switch(n int, cfg *Config) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if n < 1 {
		return fmt.Errorf("space: invalid desktop number: %d", n)
	}
	ax := newAX(cfg)
	if err := ax.Switch(n); err != nil {
		return err
	}
	settle(cfg)
	return nil
}

// List returns current Desktops (ordered by number when possible).
func List(cfg *Config) ([]Desktop, error) {
	if err := requireDarwin(); err != nil {
		return nil, err
	}
	return newAX(cfg).List()
}

// Highest returns the largest Desktop number.
func Highest(cfg *Config) (int, error) {
	if err := requireDarwin(); err != nil {
		return 0, err
	}
	return newAX(cfg).Highest()
}

// CreateAndActivate creates a Desktop, switches to it, and returns its number.
// Used when a caller needs the new Space frontmost before further work.
func CreateAndActivate(cfg *Config) (int, error) {
	if err := requireDarwin(); err != nil {
		return 0, err
	}
	ax := newAX(cfg)
	if err := ax.Create(); err != nil {
		return 0, err
	}
	n, err := ax.Highest()
	if err != nil {
		return 0, fmt.Errorf("space: created but could not resolve new desktop: %w", err)
	}
	if err := ax.Switch(n); err != nil {
		return 0, err
	}
	settle(cfg)
	return n, nil
}

func requireDarwin() error {
	if effectiveGOOS() != "darwin" {
		return ErrUnsupportedPlatform
	}
	return nil
}

func settle(cfg *Config) {
	d := defaultSettle
	if cfg != nil {
		if cfg.Settle < 0 {
			return
		}
		if cfg.Settle > 0 {
			d = cfg.Settle
		}
	}
	if d > 0 {
		time.Sleep(d)
	}
}

type axClient struct {
	osascript func(script string, args ...string) (string, error)
}

func newAX(cfg *Config) *axClient {
	c := &axClient{osascript: defaultOSAscript}
	if cfg != nil && cfg.Osascript != nil {
		c.osascript = cfg.Osascript
	}
	return c
}

func (a *axClient) Create() error {
	out, err := a.osascript(scriptCreate)
	if err != nil {
		return err
	}
	if strings.HasPrefix(out, "FAIL:") {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func (a *axClient) Switch(n int) error {
	out, err := a.osascript(scriptSwitch, strconv.Itoa(n))
	if err != nil {
		return err
	}
	if strings.HasPrefix(out, "FAIL:") {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func (a *axClient) List() ([]Desktop, error) {
	out, err := a.osascript(scriptList)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(out, "FAIL:") {
		return nil, fmt.Errorf("%s", out)
	}
	desktops := parseListOutput(out)
	if len(desktops) == 0 && !strings.Contains(out, "count=0") {
		return nil, fmt.Errorf("space: could not parse desktop list: %s", out)
	}
	return desktops, nil
}

func (a *axClient) Highest() (int, error) {
	out, err := a.osascript(scriptHighest)
	if err != nil {
		return 0, err
	}
	out = strings.TrimSpace(out)
	if strings.HasPrefix(out, "FAIL:") {
		return 0, fmt.Errorf("%s", out)
	}
	n, err := strconv.Atoi(out)
	if err != nil {
		return 0, fmt.Errorf("space: highest desktop: %w (raw=%q)", err, out)
	}
	if n < 1 {
		return 0, fmt.Errorf("space: no Desktop buttons found")
	}
	return n, nil
}

func defaultOSAscript(source string, args ...string) (string, error) {
	cmd := exec.Command("osascript", append([]string{"-"}, args...)...)
	cmd.Stdin = strings.NewReader(source)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return text, fmt.Errorf("%s", text)
		}
		return "", fmt.Errorf("osascript: %w", err)
	}
	return text, nil
}

// ParseListOutput parses scriptList stdout (exported for tests).
func ParseListOutput(out string) []Desktop {
	return parseListOutput(out)
}

func parseListOutput(out string) []Desktop {
	const prefix = "desktops=["
	i := strings.Index(out, prefix)
	if i < 0 {
		return nil
	}
	rest := out[i+len(prefix):]
	j := strings.Index(rest, "]")
	if j < 0 {
		return nil
	}
	inner := strings.TrimSpace(rest[:j])
	if inner == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	var desktops []Desktop
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		numStr := strings.TrimPrefix(p, "Desktop ")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		desktops = append(desktops, Desktop{Number: n, Name: p})
	}
	return desktops
}
