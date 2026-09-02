package rg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/os/exectry"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/binaryversion"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

const (
	// EnvRgBin is an optional absolute rg path override considered as one candidate.
	EnvRgBin = "RG_BIN"

	viaEnv        = "env"
	viaWellKnown  = "well_known"
	viaLoginShell = "login_shell"
)

// CLI is one discovered working rg executable.
type CLI struct {
	Path    string
	Version string
	Via     string
}

// WellKnownPathsOpts configures WellKnownPaths.
type WellKnownPathsOpts struct {
	Home string
}

// WellKnownPaths returns ordered candidate rg locations.
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
			filepath.Join(home, ".local", "bin", "rg"),
			filepath.Join(home, "go", "bin", "rg"),
		)
	}
	out = append(out,
		"/opt/homebrew/bin/rg",
		"/usr/local/bin/rg",
	)
	return out
}

// DiscoverOpts configures Found / Newest.
type DiscoverOpts struct {
	Home         string
	Getenv       func(string) string
	IsExecutable func(string) bool
	RunVersion   func(ctx context.Context, bin string) (string, error)
	RunLogin     func(shell, command string, env []string) (stdout string, err error)
	Timeout      time.Duration
}

// Version runs `bin --version` and returns the first X.Y.Z semver.
func Version(ctx context.Context, bin string) (string, error) {
	if strings.TrimSpace(bin) == "" {
		return "", fmt.Errorf("rg version: empty bin")
	}
	raw, err := defaultRunVersion(ctx, bin)
	if err != nil {
		return "", err
	}
	return binaryversion.ParseSemver(string(raw))
}

func defaultRunVersion(ctx context.Context, bin string) ([]byte, error) {
	return exectry.Output(ctx, bin, "--version")
}

// Found returns every unique working rg from RG_BIN, well-known paths, and
// login-shell PATH lookups. Missing / non-executable / unversioned paths are
// dropped. Zero hits is ([], nil).
func Found(ctx context.Context, opts DiscoverOpts) ([]CLI, error) {
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
		runVersion = func(ctx context.Context, bin string) (string, error) {
			raw, err := defaultRunVersion(ctx, bin)
			if err != nil {
				return "", err
			}
			return binaryversion.ParseSemver(string(raw))
		}
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

	if p := strings.TrimSpace(getenv(EnvRgBin)); p != "" {
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
	if hits, err := lookpath.LookupInAllShellPATHs("rg", lpOpts); err == nil {
		for _, p := range hits {
			add(p, viaLoginShell)
		}
	}

	versioned := binaryversion.Find(ctx, cands, func(ctx context.Context, path string) (string, error) {
		if !isExec(path) {
			return "", fmt.Errorf("rg candidate is not executable: %s", path)
		}
		return runVersion(ctx, path)
	})
	out := make([]CLI, 0, len(versioned))
	for _, c := range versioned {
		out = append(out, CLI{Path: c.Path, Version: c.Version, Via: c.Via})
	}
	return out, nil
}

// Newest returns the highest-version CLI from Found. None → error.
func Newest(ctx context.Context, opts DiscoverOpts) (CLI, error) {
	found, err := Found(ctx, opts)
	if err != nil {
		return CLI{}, err
	}
	if len(found) == 0 {
		return CLI{}, fmt.Errorf("newest rg: none found")
	}
	candidates := make([]binaryversion.Candidate, 0, len(found))
	versions := make(map[string]string, len(found))
	for _, c := range found {
		candidates = append(candidates, binaryversion.Candidate{Path: c.Path, Via: c.Via})
		versions[c.Path] = c.Version
	}
	best, err := binaryversion.Newest(ctx, candidates, func(_ context.Context, path string) (string, error) {
		return versions[path], nil
	})
	if err != nil {
		return CLI{}, err
	}
	return CLI{Path: best.Path, Version: best.Version, Via: best.Via}, nil
}

// FormatUsingNotice builds the gray-notice body (without the "notice: " prefix):
//
//	using rg 15.2.0 (/opt/homebrew/bin/rg); also found 14.1.1 (~/.local/bin/rg)
func FormatUsingNotice(selected CLI, others []CLI) string {
	msg := fmt.Sprintf("using rg %s (%s)", selected.Version, selected.Path)
	var also []string
	for _, o := range others {
		if o.Path == selected.Path {
			continue
		}
		also = append(also, fmt.Sprintf("%s (%s)", o.Version, o.Path))
	}
	if len(also) > 0 {
		msg += "; also found " + strings.Join(also, ", ")
	}
	return msg
}
