// Package lookpath resolves CLI binary names when process PATH is thin
// (Launch Services / menu-bar apps).
package lookpath

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Result is a successful resolution of a binary name.
type Result struct {
	Path string // absolute path to invocable binary
	Via  string // direct | path | extra_dir | default_dir | candidate | login_shell:<shell>
}

// Options configures Look / LookPath. Injectables are nil for production defaults.
type Options struct {
	Home            string
	ExtraDirs       []string
	ExtraCandidates []string
	Shells          []string      // default {"bash","zsh"}
	Timeout         time.Duration // default ~5s per shell

	// Injectables (nil = production)
	LookPath     func(file string) (string, error)
	IsExecutable func(path string) bool
	RunLogin     func(shell, command string, env []string) (stdout string, err error)
}

const (
	viaDirect      = "direct"
	viaPath        = "path"
	viaExtraDir    = "extra_dir"
	viaDefaultDir  = "default_dir"
	viaCandidate   = "candidate"
	viaLoginPrefix = "login_shell:"

	defaultTimeout = 5 * time.Second
)

// Look resolves name using the ordered multi-stage pipeline.
// It never mutates process env or cwd.
func Look(name string, opts Options) (Result, error) {
	if name == "" {
		return Result{}, fmt.Errorf("lookpath: empty name")
	}

	isExec := opts.IsExecutable
	if isExec == nil {
		isExec = IsExecutable
	}

	// 1. Direct path (absolute or contains path separator): no fallthrough.
	if isDirectPath(name) {
		if isExec(name) {
			return Result{Path: name, Via: viaDirect}, nil
		}
		return Result{}, fmt.Errorf("lookpath: %s: not executable or not found", name)
	}

	// 2. PATH lookup (injectable; default exec.LookPath)
	lookPath := opts.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if p, err := lookPath(name); err == nil && p != "" {
		return Result{Path: p, Via: viaPath}, nil
	}

	// 3. ExtraDirs
	for _, dir := range opts.ExtraDirs {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, name)
		if isExec(p) {
			return Result{Path: p, Via: viaExtraDir}, nil
		}
	}

	// 4. DefaultDirs(home)
	for _, dir := range DefaultDirs(opts.Home) {
		p := filepath.Join(dir, name)
		if isExec(p) {
			return Result{Path: p, Via: viaDefaultDir}, nil
		}
	}

	// 5. ExtraCandidates (absolute paths)
	for _, cand := range opts.ExtraCandidates {
		if cand == "" {
			continue
		}
		if isExec(cand) {
			return Result{Path: cand, Via: viaCandidate}, nil
		}
	}

	// 6. Login shells
	shells := opts.Shells
	if len(shells) == 0 {
		shells = []string{"bash", "zsh"}
	}
	runLogin := opts.RunLogin
	if runLogin == nil {
		timeout := opts.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		runLogin = defaultRunLogin(timeout)
	}
	cmd := "command -v " + name
	env := minimalLoginEnv(opts.Home)
	for _, shell := range shells {
		if shell == "" {
			continue
		}
		stdout, err := runLogin(shell, cmd, env)
		if err != nil {
			continue
		}
		p := strings.TrimSpace(stdout)
		if p == "" {
			continue
		}
		return Result{Path: p, Via: viaLoginPrefix + shell}, nil
	}

	// 7. Not found
	return Result{}, fmt.Errorf("lookpath: %s: not found", name)
}

// LookPath is a convenience wrapper that returns only the resolved path.
func LookPath(name string, opts Options) (string, error) {
	res, err := Look(name, opts)
	if err != nil {
		return "", err
	}
	return res.Path, nil
}

// DefaultDirs returns ordered default search directories.
// When home is non-empty: $HOME/.local/bin, $HOME/go/bin, then system bins.
// Empty home omits home-relative entries.
func DefaultDirs(home string) []string {
	system := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return append([]string(nil), system...)
	}
	return []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "go", "bin"),
		system[0],
		system[1],
	}
}

// IsExecutable reports whether path is an existing regular file with any
// execute bit set. Missing paths, directories, and non-executable files are false.
func IsExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	// Regular file with any execute bit (owner/group/other).
	mode := fi.Mode()
	if !mode.IsRegular() {
		// Still accept non-dir files that are executable (e.g. some specials);
		// contract focuses on regular files — directories already rejected.
		// Treat non-regular non-dir as non-executable for safety.
		return false
	}
	return mode&0o111 != 0
}

func isDirectPath(name string) bool {
	if filepath.IsAbs(name) {
		return true
	}
	// Relative form with path separator (./bin/tool, bin/tool, …).
	return strings.ContainsRune(name, filepath.Separator) || strings.Contains(name, "/")
}

func minimalLoginEnv(home string) []string {
	// Minimal env for login shell probes; never mutates process env.
	env := []string{
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	if home != "" {
		env = append(env, "HOME="+home)
	} else if h, err := os.UserHomeDir(); err == nil && h != "" {
		env = append(env, "HOME="+h)
	}
	return env
}

func defaultRunLogin(timeout time.Duration) func(shell, command string, env []string) (string, error) {
	return func(shell, command string, env []string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Login interactive: shell -lic 'command'
		cmd := exec.CommandContext(ctx, shell, "-lic", command)
		cmd.Env = env
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("login shell %s: %w", shell, err)
		}
		return stdout.String(), nil
	}
}
