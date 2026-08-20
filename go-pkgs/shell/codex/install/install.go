// Package install installs and updates the OpenAI Codex CLI.
//
// L2 library API: injectable HTTP, LookPath, RunShell, RunVersion, and
// FetchLatest for parallel-safe tests without real network or binaries.
package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/binaryversion"
)

const (
	// InstallScriptURL is the official Codex install script URL.
	InstallScriptURL = "https://chatgpt.com/codex/install.sh"

	// InstallCmd runs the official install script via curl | sh.
	InstallCmd = `curl -fsSL https://chatgpt.com/codex/install.sh | sh`

	// UpdateCmd is the default in-place update command.
	UpdateCmd = "codex update"

	// NPMLatestURL is the npm registry "latest" metadata endpoint for @openai/codex.
	NPMLatestURL = "https://registry.npmjs.org/@openai/codex/latest"
)

// LatestVersionOpts configures LatestVersion.
type LatestVersionOpts struct {
	// URL overrides NPMLatestURL when non-empty.
	URL string
	// HTTPClient is used for the request; nil → http.DefaultClient.
	HTTPClient *http.Client
}

// LocalVersionOpts configures LocalVersion.
type LocalVersionOpts struct {
	// Bin is a resolved binary path. Empty → resolve via LookPath("codex").
	Bin string
	// LookPath resolves a binary name on PATH; nil → exec.LookPath.
	LookPath func(file string) (string, error)
	// RunVersion runs the version command; nil → `bin --version`.
	RunVersion func(ctx context.Context, bin string) (string, error)
}

// InstallOpts configures Install.
type InstallOpts struct {
	// RunShell runs a shell command string; nil → sh -c.
	RunShell func(ctx context.Context, cmd string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

// UpdateOpts configures Update.
type UpdateOpts struct {
	// Bin, when set, may path-qualify the update command (optional).
	Bin string
	// RunShell runs a shell command string; nil → sh -c.
	RunShell func(ctx context.Context, cmd string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

// EnsureOpts configures Ensure.
type EnsureOpts struct {
	// Bin is a preferred resolved path passed to LookPath consumers / Update.
	Bin string
	// LookPath resolves "codex" on PATH; nil → exec.LookPath.
	LookPath func(file string) (string, error)
	// RunShell runs install/update shell commands; nil → sh -c.
	RunShell func(ctx context.Context, cmd string) error
	// RunVersion runs the local version command; nil → `bin --version`.
	RunVersion func(ctx context.Context, bin string) (string, error)
	// FetchLatest returns the latest remote version; nil → LatestVersion.
	FetchLatest func(ctx context.Context) (string, error)
	// HTTPClient is used by the default FetchLatest path.
	HTTPClient *http.Client
	Stdout     io.Writer
	Stderr     io.Writer
}

// Result is the outcome of Ensure.
type Result struct {
	// Action is one of: install | update | noop.
	Action        string
	BinPath       string
	LocalVersion  string
	LatestVersion string
	NeedsUpdate   bool
}

// ParseVersion extracts the first semver X.Y.Z from version command or package text.
// Empty input or no match returns an error.
func ParseVersion(output string) (string, error) {
	return binaryversion.ParseSemver(output)
}

// NeedsUpdate reports whether local is strictly older than latest.
// Returns false when either side is empty/unparseable, equal, or local > latest.
func NeedsUpdate(local, latest string) bool {
	cmp, err := binaryversion.CompareSemver(local, latest)
	return err == nil && cmp < 0
}

// LatestVersion GETs the npm (or override) latest metadata and returns JSON "version".
func LatestVersion(ctx context.Context, opts LatestVersionOpts) (string, error) {
	url := opts.URL
	if url == "" {
		url = NPMLatestURL
	}
	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("latest version: HTTP %d from %s", resp.StatusCode, url)
	}

	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("latest version: decode JSON: %w", err)
	}
	if body.Version == "" {
		return "", fmt.Errorf("latest version: empty version field")
	}
	return body.Version, nil
}

// LocalVersion resolves the codex binary and returns raw version-command stdout.
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
			return fmt.Errorf("codex update: %w; install: %v", err, instErr)
		}
		return nil
	}
	return nil
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

// Ensure installs or updates Codex as needed.
//
//   - missing bin → Install; Action=install; does not fetch latest
//   - present + NeedsUpdate → Update; Action=update
//   - present + !NeedsUpdate → noop; Action=noop
//   - present + latest/local unknown → noop; Action=noop
//
// Production lookup (nil LookPath) uses NewestCodex. Injected LookPath is the
// test seam and keeps first-hit semantics.
func Ensure(ctx context.Context, opts EnsureOpts) (Result, error) {
	var result Result

	binPath, lookErr := resolveEnsureBin(ctx, opts)
	if lookErr != nil {
		// Missing → install; never fetch latest.
		if err := Install(ctx, InstallOpts{
			RunShell: opts.RunShell,
			Stdout:   opts.Stdout,
			Stderr:   opts.Stderr,
		}); err != nil {
			result.Action = "install"
			return result, err
		}
		result.Action = "install"
		if opts.Bin != "" {
			result.BinPath = opts.Bin
		}
		return result, nil
	}

	result.BinPath = binPath
	if result.BinPath == "" && opts.Bin != "" {
		result.BinPath = opts.Bin
	}

	// Present → local version (soft on failure).
	runVersion := opts.RunVersion
	if runVersion == nil {
		runVersion = defaultRunVersion
	}
	rawLocal, localErr := runVersion(ctx, binPath)
	if localErr == nil {
		if parsed, err := ParseVersion(rawLocal); err == nil {
			result.LocalVersion = parsed
		} else {
			result.LocalVersion = rawLocal
		}
	}

	// Fetch latest only when bin is present.
	fetchLatest := opts.FetchLatest
	if fetchLatest == nil {
		fetchLatest = func(ctx context.Context) (string, error) {
			return LatestVersion(ctx, LatestVersionOpts{HTTPClient: opts.HTTPClient})
		}
	}
	latest, latestErr := fetchLatest(ctx)
	if latestErr != nil {
		// Unknown latest → noop (no install/update); prefer soft nil error.
		result.Action = "noop"
		return result, nil
	}
	result.LatestVersion = latest

	localForCompare := result.LocalVersion
	if localForCompare == "" && localErr == nil {
		localForCompare = rawLocal
	}
	// If local version command failed, treat as unparseable → no forced update.
	if localErr != nil {
		result.Action = "noop"
		result.NeedsUpdate = false
		return result, nil
	}

	result.NeedsUpdate = NeedsUpdate(localForCompare, latest)
	if !result.NeedsUpdate {
		result.Action = "noop"
		return result, nil
	}

	if err := Update(ctx, UpdateOpts{
		Bin:      binPath,
		RunShell: opts.RunShell,
		Stdout:   opts.Stdout,
		Stderr:   opts.Stderr,
	}); err != nil {
		result.Action = "update"
		return result, err
	}
	result.Action = "update"
	return result, nil
}

func resolveEnsureBin(ctx context.Context, opts EnsureOpts) (string, error) {
	if opts.LookPath != nil {
		return opts.LookPath("codex")
	}
	p, _, err := NewestCodex(ctx, NewestCodexOpts{
		RunVersion: opts.RunVersion,
	})
	return p, err
}

func resolveBin(bin string, lookPath func(file string) (string, error)) (string, error) {
	if bin != "" {
		return bin, nil
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	return lookPath("codex")
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
