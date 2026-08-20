package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/binaryversion"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

const (
	// EnvCodexBin is an optional absolute Codex path override considered as one
	// discovery candidate (not a sticky winner).
	EnvCodexBin = "CODEX_BIN"

	viaEnv        = "env"
	viaWellKnown  = "well_known"
	viaLoginShell = "login_shell"
)

// WellKnownPathsOpts configures WellKnownPaths.
type WellKnownPathsOpts struct {
	// Home is used for $home/… entries. Empty → os.UserHomeDir().
	// Still empty after that → system dirs only.
	Home string
}

// NewestCodexOpts configures FoundCodex / NewestCodex. Nil injectables use
// production defaults. Never mutates process env or cwd.
type NewestCodexOpts struct {
	Home         string
	Getenv       func(string) string
	IsExecutable func(string) bool
	RunVersion   func(ctx context.Context, bin string) (string, error)
	RunLogin     func(shell, command string, env []string) (stdout string, err error)
	Timeout      time.Duration
}

// CodexCLI is one discovered working Codex executable.
type CodexCLI struct {
	Path    string // filepath.Clean absolute (or cleaned join)
	Version string // parsed X.Y.Z
	Via     string // env | well_known | login_shell
}

// WellKnownPaths returns the ordered candidate Codex locations.
func WellKnownPaths(opts WellKnownPathsOpts) []string {
	home := strings.TrimSpace(opts.Home)
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = strings.TrimSpace(h)
		}
	}
	var out []string
	if home != "" {
		out = append(out,
			filepath.Join(home, ".local", "bin", "codex"),
			filepath.Join(home, "go", "bin", "codex"),
		)
	}
	out = append(out,
		"/opt/homebrew/bin/codex",
		"/usr/local/bin/codex",
	)
	if home != "" {
		out = append(out,
			filepath.Join(home, ".npm-global", "bin", "codex"),
			filepath.Join(home, ".volta", "bin", "codex"),
		)
	}
	return out
}

// Version runs `bin --version` and returns the parsed X.Y.Z.
// Empty bin, spawn failure, or no semver is an error.
func Version(ctx context.Context, bin string) (string, error) {
	if strings.TrimSpace(bin) == "" {
		return "", fmt.Errorf("codex version: empty bin")
	}
	raw, err := defaultRunVersion(ctx, bin)
	if err != nil {
		return "", err
	}
	return ParseVersion(raw)
}

// FoundCodex returns every unique working Codex CLI from CODEX_BIN, well-known
// paths, and login-shell PATH lookups. Missing, non-executable, and unversioned
// paths are dropped. Zero hits is ([], nil). Login dump failure is skipped
// (not fatal) so well-known hits still return.
func FoundCodex(ctx context.Context, opts NewestCodexOpts) ([]CodexCLI, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	isExec := opts.IsExecutable
	if isExec == nil {
		isExec = lookpath.IsExecutable
	}
	getenv := opts.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	runVersion := opts.RunVersion
	if runVersion == nil {
		runVersion = defaultRunVersion
	}

	home := strings.TrimSpace(opts.Home)
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = strings.TrimSpace(h)
		}
	}

	var cands []binaryversion.Candidate
	seen := make(map[string]struct{})
	add := func(p, via string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		cleaned := filepath.Clean(p)
		if _, ok := seen[cleaned]; ok {
			return
		}
		seen[cleaned] = struct{}{}
		cands = append(cands, binaryversion.Candidate{Path: cleaned, Via: via})
	}

	if p := strings.TrimSpace(getenv(EnvCodexBin)); p != "" {
		add(p, viaEnv)
	}
	for _, p := range WellKnownPaths(WellKnownPathsOpts{Home: home}) {
		add(p, viaWellKnown)
	}

	lpOpts := lookpath.Options{
		Home:         home,
		Timeout:      opts.Timeout,
		RunLogin:     opts.RunLogin,
		IsExecutable: isExec,
	}
	if hits, err := lookpath.LookupInAllShellPATHs("codex", lpOpts); err == nil {
		for _, p := range hits {
			add(p, viaLoginShell)
		}
	}

	versioned := binaryversion.Find(ctx, cands, func(ctx context.Context, path string) (string, error) {
		if !isExec(path) {
			return "", fmt.Errorf("codex candidate is not executable: %s", path)
		}
		return runVersion(ctx, path)
	})
	out := make([]CodexCLI, 0, len(versioned))
	for _, c := range versioned {
		out = append(out, CodexCLI{Path: c.Path, Version: c.Version, Via: c.Via})
	}
	return out, nil
}

// NewestCodex returns the highest-version row from FoundCodex.
// Version ties keep list order (first wins). None parseable → error.
func NewestCodex(ctx context.Context, opts NewestCodexOpts) (binPath, version string, err error) {
	found, err := FoundCodex(ctx, opts)
	if err != nil {
		return "", "", err
	}
	if len(found) == 0 {
		return "", "", fmt.Errorf("newest codex: none found")
	}
	candidates := make([]binaryversion.Candidate, 0, len(found))
	for _, c := range found {
		candidates = append(candidates, binaryversion.Candidate{Path: c.Path, Via: c.Via})
	}
	versions := make(map[string]string, len(found))
	for _, c := range found {
		versions[c.Path] = c.Version
	}
	best, err := binaryversion.Newest(ctx, candidates, func(_ context.Context, path string) (string, error) {
		return versions[path], nil
	})
	if err != nil {
		return "", "", err
	}
	return best.Path, best.Version, nil
}
