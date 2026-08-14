// Package chrome provides macOS UI automation to load an unpacked Chrome
// extension (Developer mode → Load unpacked → folder).
//
// Behavior is a Go port of the rule-driven script at
// working/ai/computer-use/mlx-use-example/load_chrome_extension.py (no LLM).
//
// A Swift port of the same steps (in-process for a GUI host) lives in
// swift/ChromeLoadUnpacked.swift. Marcus.app copies that file; edit the Swift
// there, not the Marcus tree.
//
// Requirements (darwin):
//   - Google Chrome installed
//   - Accessibility trust for the process running this package
//   - Often Automation permission for Google Chrome / System Events
//
// Non-darwin builds return ErrUnsupported.
package chrome

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExtensionsURL is the Chrome extensions management page.
const ExtensionsURL = "chrome://extensions"

// DefaultAppName is the localized Google Chrome app name on macOS.
const DefaultAppName = "Google Chrome"

// ErrUnsupported is returned when LoadUnpacked is called on a non-macOS build
// or when the platform cannot drive Chrome UI.
var ErrUnsupported = errors.New("chrome: LoadUnpacked is only supported on macOS")

// LoadUnpackedOpts configures a LoadUnpacked run.
type LoadUnpackedOpts struct {
	// ExtensionDir is the unpacked extension folder (must contain manifest.json).
	ExtensionDir string
	// AppName defaults to DefaultAppName ("Google Chrome").
	AppName string
	// DryRun opens chrome://extensions and reports controls; does not click Load unpacked.
	DryRun bool
	// DumpTree prints a filtered Accessibility/UI tree to Stderr and returns
	// without loading (still opens the extensions page).
	DumpTree bool
	// VerifyName is optional text expected on the page after load (default "Browser Agent").
	VerifyName string
	// VerifyVersion is the version string of the just-loaded package (optional).
	// Used to keep that card when removing older same-name extensions.
	// When empty, inferred from the last path segment of ExtensionDir when it looks version-like.
	VerifyVersion string
	// RemoveOlder removes other extension cards with the same VerifyName after a
	// successful load (Loaded or SubmittedUnknown). nil means true (default on).
	// Set to PtrBool(false) or use KeepOlder to skip cleanup.
	RemoveOlder *bool
	// KeepOlder when true forces RemoveOlder off (CLI --keep-old).
	// Overrides RemoveOlder when true.
	KeepOlder bool
	// PageTimeout waits for the extensions UI (default 20s).
	PageTimeout time.Duration
	// DialogTimeout waits for the open-folder sheet (default 10s).
	DialogTimeout time.Duration
	// VerifyTimeout waits for the extension card after load (default 8s).
	VerifyTimeout time.Duration
	// Stdout receives step lines (optional; defaults to os.Stdout).
	Stdout io.Writer
	// Stderr receives warnings, dump-tree, and multi-instance notes (optional; defaults to os.Stderr).
	Stderr io.Writer
	// ScreenshotDir when non-empty saves a PNG after each major step (macOS screencapture).
	// Used for human debugging of the UI load flow.
	ScreenshotDir string
}

// LoadUnpackedResult may include screenshot paths when ScreenshotDir is set.

// LoadUnpackedResult summarizes a run.
type LoadUnpackedResult struct {
	// DeveloperModeVisible is true when the Developer mode control was found.
	DeveloperModeVisible bool
	// LoadUnpackedVisible is true when the Load unpacked button was found.
	LoadUnpackedVisible bool
	// ExtensionListed is true when VerifyName appeared on the page (best-effort).
	ExtensionListed bool
	// Loaded is true when a full load completed and verify succeeded.
	Loaded bool
	// SubmittedUnknown is true when folder pick ran but verify could not confirm.
	SubmittedUnknown bool
	// MultiInstanceWarned is true when more than one Chrome user-data-dir was detected.
	MultiInstanceWarned bool
	// RemovedOlder is how many same-name extension cards were removed (best-effort).
	RemovedOlder int
	// RemoveOlderAttempted is true when post-load cleanup ran.
	RemoveOlderAttempted bool
	// Screenshots lists paths written when ScreenshotDir is set (ordered by step).
	Screenshots []string
}

// PtrBool returns a *bool for LoadUnpackedOpts.RemoveOlder.
func PtrBool(b bool) *bool { return &b }

// removeOlderEnabled reports whether post-load cleanup should run.
// Default true; KeepOlder or RemoveOlder==false disables.
func removeOlderEnabled(opts LoadUnpackedOpts) bool {
	if opts.KeepOlder {
		return false
	}
	if opts.RemoveOlder == nil {
		return true
	}
	return *opts.RemoveOlder
}

// isTimeoutErr reports context/command timeouts from runOSAscript.
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// exec.CommandContext often surfaces "signal: killed" / deadline text.
	msg := err.Error()
	return strings.Contains(msg, "context deadline exceeded") ||
		strings.Contains(msg, "signal: killed")
}

// InferVersionFromDir returns the last path segment when it looks like a version
// (digits and dots), else "".
func InferVersionFromDir(extensionDir string) string {
	base := filepath.Base(strings.TrimRight(strings.TrimSpace(extensionDir), string(filepath.Separator)))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	// version-like: starts with digit, only [0-9.]
	if base[0] < '0' || base[0] > '9' {
		return ""
	}
	for _, r := range base {
		if (r < '0' || r > '9') && r != '.' {
			return ""
		}
	}
	return base
}

// ExtensionDirOK reports whether path looks like an unpacked Chrome extension.
func ExtensionDirOK(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		return false
	}
	fi, err := os.Stat(filepath.Join(path, "manifest.json"))
	return err == nil && !fi.IsDir()
}

// LoadUnpacked drives Chrome UI to load (or inspect) an unpacked extension.
// On non-darwin platforms it returns ErrUnsupported.
func LoadUnpacked(ctx context.Context, opts LoadUnpackedOpts) (LoadUnpackedResult, error) {
	return loadUnpacked(ctx, opts)
}

func normalizeOpts(opts LoadUnpackedOpts) (LoadUnpackedOpts, error) {
	if strings.TrimSpace(opts.ExtensionDir) == "" {
		return opts, fmt.Errorf("chrome: ExtensionDir is required")
	}
	abs, err := filepath.Abs(opts.ExtensionDir)
	if err != nil {
		return opts, fmt.Errorf("chrome: ExtensionDir: %w", err)
	}
	opts.ExtensionDir = abs
	if !ExtensionDirOK(opts.ExtensionDir) {
		return opts, fmt.Errorf("chrome: extension dir missing or no manifest.json: %s", opts.ExtensionDir)
	}
	if strings.TrimSpace(opts.AppName) == "" {
		opts.AppName = DefaultAppName
	}
	if strings.TrimSpace(opts.VerifyName) == "" {
		opts.VerifyName = "Browser Agent"
	}
	if strings.TrimSpace(opts.VerifyVersion) == "" {
		opts.VerifyVersion = InferVersionFromDir(opts.ExtensionDir)
	}
	if opts.PageTimeout <= 0 {
		opts.PageTimeout = 20 * time.Second
	}
	if opts.DialogTimeout <= 0 {
		opts.DialogTimeout = 10 * time.Second
	}
	if opts.VerifyTimeout <= 0 {
		opts.VerifyTimeout = 8 * time.Second
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	return opts, nil
}

func stepf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, format, args...)
	if !strings.HasSuffix(format, "\n") {
		_, _ = io.WriteString(w, "\n")
	}
}

// escapeAS escapes a string for embedding in an AppleScript double-quoted literal.
func escapeAS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
