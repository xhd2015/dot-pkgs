package lookpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AllShellPATHs returns unique login-shell PATH directories from bash then zsh.
// Equivalent to AllShellPATHsWith(Options{}).
func AllShellPATHs() ([]string, error) {
	return AllShellPATHsWith(Options{})
}

// AllShellPATHsWith returns unique directories from login bash PATH, then
// zsh-only PATH entries (first-seen). One shell error or empty PATH is skipped.
// Both shells failing returns the last error. Process PATH is never read.
// Never mutates process env or cwd.
func AllShellPATHsWith(opts Options) ([]string, error) {
	loginOpts := LoginEnvOptions{
		Home:     opts.Home,
		Timeout:  opts.Timeout,
		RunLogin: opts.RunLogin,
	}

	var dirs []string
	seen := make(map[string]struct{})
	var lastErr error
	anyOK := false

	for _, resolve := range []func(string, LoginEnvOptions) (string, error){
		ResolveBashLoginEnv,
		ResolveZshLoginEnv,
	} {
		p, err := resolve("PATH", loginOpts)
		if err != nil {
			lastErr = err
			continue
		}
		anyOK = true
		for _, dir := range splitPathList(p) {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}
			cleaned := filepath.Clean(dir)
			if cleaned == "." && dir != "." {
				continue
			}
			if _, ok := seen[cleaned]; ok {
				continue
			}
			seen[cleaned] = struct{}{}
			dirs = append(dirs, cleaned)
		}
	}

	if !anyOK {
		if lastErr != nil {
			return nil, lastErr
		}
		return []string{}, nil
	}
	if dirs == nil {
		return []string{}, nil
	}
	return dirs, nil
}

// LookupInAllShellPATHs returns every executable named `name` under
// AllShellPATHsWith directories (which -a semantics without the which binary).
// Empty name is an error. Zero hits is ([], nil). AllShellPATHs errors propagate.
func LookupInAllShellPATHs(name string, opts Options) ([]string, error) {
	if name == "" {
		return nil, fmt.Errorf("lookpath: empty name")
	}
	dirs, err := AllShellPATHsWith(opts)
	if err != nil {
		return nil, err
	}
	isExec := opts.IsExecutable
	if isExec == nil {
		isExec = IsExecutable
	}
	var out []string
	seen := make(map[string]struct{})
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		p := filepath.Clean(filepath.Join(dir, name))
		if !isExec(p) {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if out == nil {
		return []string{}, nil
	}
	return out, nil
}

func splitPathList(p string) []string {
	if p == "" {
		return nil
	}
	return strings.Split(p, string(os.PathListSeparator))
}
