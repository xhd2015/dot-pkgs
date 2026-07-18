// Package space automates macOS Mission Control Desktops (Spaces) via
// Accessibility / AppleScript. It does not parse CLI flags or run follow-up
// commands — callers (e.g. kool macos space) own that.
//
// Create returns ErrMaxDesktops (errors.Is) when macOS is already at the
// 16-Desktop hard limit. CountDesktops optionally counts Desktops without
// opening Mission Control (default: com.apple.spaces plist); it is not used
// by Create automatically.
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
	// ErrMaxDesktops is returned when Create cannot add a Desktop because
	// macOS is already at its hard maximum (16 Mission Control Desktops).
	// Callers can detect it with errors.Is(err, space.ErrMaxDesktops).
	ErrMaxDesktops = errors.New("space: already at macOS maximum of 16 Desktops")
)

const (
	defaultSettle = 500 * time.Millisecond
	// postCreateSettle waits after Create before Highest/Switch so the Spaces
	// Bar AX tree includes the new Desktop (without this, Switch often fails
	// with "desktop not found: Desktop N").
	postCreateSettle = 400 * time.Millisecond
	// switchRetries is how many times to re-query Highest + Switch after Create
	// (does not re-Create — that would spawn extra Desktops near the 16 cap).
	switchRetries = 3
	// createRetries for flaky Mission Control open (-1719 Invalid index).
	createRetries = 3
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
//
// Stability notes (Mission Control AX is racy):
//   - Create is retried on transient Dock/MC errors (-1719 Invalid index).
//   - After a successful Create, settle briefly, then Highest+Switch with
//     retries. Switch failures do NOT re-Create (avoids extra Desktops).
func CreateAndActivate(cfg *Config) (int, error) {
	if err := requireDarwin(); err != nil {
		return 0, err
	}
	ax := newAX(cfg)

	var createErr error
	for attempt := 0; attempt < createRetries; attempt++ {
		createErr = ax.Create()
		if createErr == nil {
			break
		}
		// Capacity errors are permanent; do not retry Create.
		if errors.Is(createErr, ErrMaxDesktops) || !isTransientSpaceError(createErr) || attempt+1 == createRetries {
			return 0, createErr
		}
		settle(cfg)
	}

	// Let the new Desktop appear in the Spaces Bar before Highest/Switch.
	sleepDuration(cfg, postCreateSettle)

	var lastErr error
	for attempt := 0; attempt < switchRetries; attempt++ {
		n, err := ax.Highest()
		if err != nil {
			lastErr = fmt.Errorf("space: created but could not resolve new desktop: %w", err)
			if !isTransientSpaceError(err) || attempt+1 == switchRetries {
				return 0, lastErr
			}
			settle(cfg)
			continue
		}
		if err := ax.Switch(n); err != nil {
			lastErr = err
			if !isTransientSpaceError(err) || attempt+1 == switchRetries {
				return 0, lastErr
			}
			settle(cfg)
			continue
		}
		settle(cfg)
		return n, nil
	}
	if lastErr != nil {
		return 0, lastErr
	}
	return 0, fmt.Errorf("space: CreateAndActivate failed after retries")
}

// isTransientSpaceError reports Mission Control AX races worth retrying.
func isTransientSpaceError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "desktop not found"):
		return true
	case strings.Contains(s, "Invalid index"):
		return true
	case strings.Contains(s, "-1719"):
		return true
	case strings.Contains(s, "no Desktop buttons found"):
		return true
	case strings.Contains(s, "can’t get group") || strings.Contains(s, "Can't get group"):
		return true
	case strings.Contains(s, "Mission Control"):
		// Broad: MC group/list missing while animation runs.
		return strings.Contains(s, "FAIL:")
	default:
		return false
	}
}

// sleepDuration sleeps d unless cfg.Settle is negative (tests disable sleeps).
func sleepDuration(cfg *Config, d time.Duration) {
	if cfg != nil && cfg.Settle < 0 {
		return
	}
	if d > 0 {
		time.Sleep(d)
	}
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
		return mapCreateError(err)
	}
	if strings.HasPrefix(out, "FAIL:") {
		return mapCreateError(fmt.Errorf("%s", out))
	}
	return nil
}

// mapCreateError classifies Mission Control create failures.
// Max-cap messages from scriptCreate become ErrMaxDesktops (errors.Is).
func mapCreateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrMaxDesktops) {
		return err
	}
	if isMaxDesktopsMessage(err.Error()) {
		return fmt.Errorf("%w", ErrMaxDesktops)
	}
	return err
}

// isMaxDesktopsMessage reports AppleScript / osascript text for the 16-Desktop cap.
func isMaxDesktopsMessage(s string) bool {
	// scriptCreate: "FAIL: cannot create Desktop: already at macOS maximum of 16 ..."
	return strings.Contains(s, "already at macOS maximum of 16") ||
		strings.Contains(s, "macOS maximum of 16")
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
