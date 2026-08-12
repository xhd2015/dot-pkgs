package lookpath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoPathOptions configures ResolveGoPath / ResolveGoPathWith.
// Injectables are nil for production defaults.
// Never mutates process env or cwd.
type GoPathOptions struct {
	// Login shell discovery (Home, Timeout, RunLogin, ShellBin)
	LoginEnv LoginEnvOptions
	// LookPath for "go"; nil → lookpath.LookPath via Options
	LookPath func(file string) (string, error)
	// RunGoEnv(goBin) returns stdout of `go env GOPATH` (caller trims)
	RunGoEnv func(goBin string) (string, error)
	// Home for ~/go fallback; empty → LoginEnv.Home then UserHomeDir
	Home string
}

// ResolveGoPath resolves GOPATH with production defaults.
func ResolveGoPath() (string, error) {
	return ResolveGoPathWith(GoPathOptions{})
}

// ResolveGoPathWith resolves a usable GOPATH via cascade:
//
//  1. Bash login GOPATH (soft empty/error)
//  2. Zsh login GOPATH (soft empty/error)
//  3. LookPath("go") + RunGoEnv / go env GOPATH (soft miss/error/empty)
//  4. filepath.Join(home, "go") — hard error only if home unresolvable
//
// Multi-GOPATH values yield filepath.Clean of the first PathListSeparator segment.
// Never mutates process env or cwd. Process os.Getenv("GOPATH") is not used.
func ResolveGoPathWith(opts GoPathOptions) (string, error) {
	loginOpts := opts.LoginEnv
	if loginOpts.Home == "" && opts.Home != "" {
		loginOpts.Home = opts.Home
	}

	// 1. Bash login GOPATH
	if p, err := ResolveBashLoginEnv("GOPATH", loginOpts); err == nil {
		if seg := firstGoPathSegment(p); seg != "" {
			return seg, nil
		}
	}

	// 2. Zsh login GOPATH
	if p, err := ResolveZshLoginEnv("GOPATH", loginOpts); err == nil {
		if seg := firstGoPathSegment(p); seg != "" {
			return seg, nil
		}
	}

	// 3. go binary + go env GOPATH
	home, homeErr := resolveGoPathHome(opts)
	lookPath := opts.LookPath
	if lookPath == nil {
		lookPath = func(file string) (string, error) {
			return LookPath(file, Options{
				Home:     firstNonEmpty(opts.Home, loginOpts.Home),
				Timeout:  loginOpts.Timeout,
				RunLogin: loginOpts.RunLogin,
			})
		}
	}
	runGoEnv := opts.RunGoEnv
	if runGoEnv == nil {
		runGoEnv = defaultRunGoEnv
	}

	if goBin, err := lookPath("go"); err == nil && goBin != "" {
		if out, err := runGoEnv(goBin); err == nil {
			if seg := firstGoPathSegment(out); seg != "" {
				return seg, nil
			}
		}
	}

	// 4. ~/go fallback
	if homeErr != nil {
		return "", homeErr
	}
	return filepath.Join(home, "go"), nil
}

func resolveGoPathHome(opts GoPathOptions) (string, error) {
	home := strings.TrimSpace(opts.Home)
	if home == "" {
		home = strings.TrimSpace(opts.LoginEnv.Home)
	}
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("lookpath: resolve home: %w", err)
		}
		home = strings.TrimSpace(h)
	}
	if home == "" {
		return "", fmt.Errorf("lookpath: empty home")
	}
	return home, nil
}

// firstGoPathSegment returns filepath.Clean of the first PathListSeparator
// segment after TrimSpace. Empty input yields "".
func firstGoPathSegment(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.IndexByte(raw, os.PathListSeparator); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return filepath.Clean(raw)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func defaultRunGoEnv(goBin string) (string, error) {
	cmd := exec.Command(goBin, "env", "GOPATH")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("lookpath: go env GOPATH: %w", err)
	}
	return string(out), nil
}
