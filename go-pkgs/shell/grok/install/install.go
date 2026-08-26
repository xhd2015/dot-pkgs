// Package install installs and updates the xAI Grok CLI.
//
// Version/upgrade SSOT is `grok update --check --json` (channel-aware; do not
// pass --alpha/--stable on check — those mutate config). L2 library API:
// injectable RunCheck / RunShell / LookPath for tests without real network.
package install

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/binaryversion"
)

const (
	// InstallScriptURL is the official Grok install script URL.
	InstallScriptURL = "https://x.ai/cli/install.sh"

	// InstallCmd runs the official install script via curl | bash.
	InstallCmd = `curl -fsSL https://x.ai/cli/install.sh | bash`

	// UpdateCmd is the default in-place update command.
	UpdateCmd = "grok update"
)

// CheckResult is the parsed `grok update --check --json` payload.
type CheckResult struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Installer       string `json:"installer"`
	Channel         string `json:"channel"`
	// AutoUpdate is left untyped in product use (null or object); ignored here.
	Error *string `json:"error"`
}

// CheckUpdateOpts configures CheckUpdate.
type CheckUpdateOpts struct {
	// Bin is a resolved binary path. Empty → resolve via LookPath("grok").
	Bin string
	// LookPath resolves a binary name on PATH; nil → exec.LookPath.
	LookPath func(file string) (string, error)
	// RunCheck runs `bin update --check --json` and returns combined stdout
	// (and optionally stderr). nil → exec the binary.
	RunCheck func(ctx context.Context, bin string) (string, error)
}

// LocalVersionOpts configures LocalVersion.
type LocalVersionOpts struct {
	Bin      string
	LookPath func(file string) (string, error)
	// RunVersion runs the version command; nil → `bin --version`.
	RunVersion func(ctx context.Context, bin string) (string, error)
}

// InstallOpts configures Install.
type InstallOpts struct {
	RunShell func(ctx context.Context, cmd string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

// UpdateOpts configures Update.
type UpdateOpts struct {
	Bin      string
	RunShell func(ctx context.Context, cmd string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

// ParseCheckJSON extracts CheckResult from `grok update --check --json` output.
// Tolerates a leading non-JSON line (e.g. "Switched to … channel.") by using
// the last line that starts with '{'.
func ParseCheckJSON(output string) (CheckResult, error) {
	raw := strings.TrimSpace(output)
	if raw == "" {
		return CheckResult{}, fmt.Errorf("grok check: empty output")
	}
	line := raw
	if !strings.HasPrefix(raw, "{") {
		var last string
		for _, ln := range strings.Split(raw, "\n") {
			ln = strings.TrimSpace(ln)
			if strings.HasPrefix(ln, "{") {
				last = ln
			}
		}
		if last == "" {
			return CheckResult{}, fmt.Errorf("grok check: no JSON object in output")
		}
		line = last
	}
	var out CheckResult
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		return CheckResult{}, fmt.Errorf("grok check: decode JSON: %w", err)
	}
	if strings.TrimSpace(out.CurrentVersion) == "" && strings.TrimSpace(out.LatestVersion) == "" {
		return CheckResult{}, fmt.Errorf("grok check: missing version fields")
	}
	return out, nil
}

// ParseVersion extracts the first semver X.Y.Z from version-command text.
func ParseVersion(output string) (string, error) {
	return binaryversion.ParseSemver(output)
}

// NeedsUpdate reports whether local is strictly older than latest.
// Returns false when either side is empty/unparseable, equal, or local > latest.
func NeedsUpdate(local, latest string) bool {
	cmp, err := binaryversion.CompareSemver(local, latest)
	return err == nil && cmp < 0
}

// NeedsUpdateFromCheck reports upgrade from a CheckResult.
// Prefers semver compare when both versions parse; otherwise falls back to
// UpdateAvailable.
func NeedsUpdateFromCheck(c CheckResult) bool {
	if NeedsUpdate(c.CurrentVersion, c.LatestVersion) {
		return true
	}
	if _, err := binaryversion.ParseSemver(c.CurrentVersion); err != nil {
		return c.UpdateAvailable
	}
	if _, err := binaryversion.ParseSemver(c.LatestVersion); err != nil {
		return c.UpdateAvailable
	}
	return false
}

// CheckUpdate runs `bin update --check --json` (no channel flags) and parses it.
func CheckUpdate(ctx context.Context, opts CheckUpdateOpts) (CheckResult, error) {
	bin, err := resolveBin(opts.Bin, opts.LookPath)
	if err != nil {
		return CheckResult{}, err
	}
	run := opts.RunCheck
	if run == nil {
		run = defaultRunCheck
	}
	out, err := run(ctx, bin)
	if err != nil {
		return CheckResult{}, err
	}
	return ParseCheckJSON(out)
}

// LocalVersion resolves the grok binary and returns raw `bin --version` stdout.
func LocalVersion(ctx context.Context, opts LocalVersionOpts) (string, error) {
	bin, err := resolveBin(opts.Bin, opts.LookPath)
	if err != nil {
		return "", err
	}
	run := opts.RunVersion
	if run == nil {
		run = defaultRunVersion
	}
	return run(ctx, bin)
}

// Install runs InstallCmd via RunShell.
func Install(ctx context.Context, opts InstallOpts) error {
	run := opts.RunShell
	if run == nil {
		run = defaultRunShell(opts.Stdout, opts.Stderr)
	}
	return run(ctx, InstallCmd)
}

// Update runs a path-qualified `<bin> update` when Bin is set, otherwise UpdateCmd.
// If that command fails, it always runs InstallCmd.
func Update(ctx context.Context, opts UpdateOpts) error {
	run := opts.RunShell
	if run == nil {
		run = defaultRunShell(opts.Stdout, opts.Stderr)
	}
	cmd := updateCmdForBin(opts.Bin)
	if err := run(ctx, cmd); err != nil {
		if instErr := Install(ctx, InstallOpts{
			RunShell: run,
			Stdout:   opts.Stdout,
			Stderr:   opts.Stderr,
		}); instErr != nil {
			return fmt.Errorf("grok update: %w; install: %v", err, instErr)
		}
		return nil
	}
	return nil
}

// UpdateCmdForBin returns a shell-safe path-qualified update command.
func UpdateCmdForBin(bin string) string {
	return updateCmdForBin(bin)
}

func updateCmdForBin(bin string) string {
	bin = strings.TrimSpace(bin)
	if bin == "" {
		return UpdateCmd
	}
	return shellQuote(bin) + " update"
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func resolveBin(bin string, lookPath func(file string) (string, error)) (string, error) {
	if bin != "" {
		return bin, nil
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	return lookPath("grok")
}

func defaultRunShell(stdout, stderr io.Writer) func(ctx context.Context, cmd string) error {
	return func(ctx context.Context, cmd string) error {
		c := exec.CommandContext(ctx, "sh", "-c", cmd)
		c.Stdout = stdout
		c.Stderr = stderr
		return c.Run()
	}
}

func defaultRunVersion(ctx context.Context, bin string) (string, error) {
	c := exec.CommandContext(ctx, bin, "--version")
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func defaultRunCheck(ctx context.Context, bin string) (string, error) {
	c := exec.CommandContext(ctx, bin, "update", "--check", "--json")
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	// Prefer stdout; append stderr only when stdout has no JSON object
	// (channel-switch messages may appear on either stream).
	out := stdout.String()
	if !strings.Contains(out, "{") && stderr.Len() > 0 {
		out = strings.TrimSpace(out + "\n" + stderr.String())
	}
	if err != nil {
		if out != "" {
			return out, err
		}
		return "", err
	}
	return out, nil
}
