package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

// DefaultManagedCLIs are bare tool names resolved through a login shell when
// the host's own PATH lacks them. A GUI-launched server usually inherits a
// minimal PATH, so a managed service that runs `go`, `node` or `npm` would
// otherwise fail even though the user's terminal finds them.
var DefaultManagedCLIs = []string{"go", "node", "npm"}

// EnvOptions describes the environment layers a managed process gets on top of
// the host's environment.
type EnvOptions struct {
	// Base is the starting environment (os.Environ style). Nil means
	// os.Environ(). Hosts that enrich PATH themselves pass their result here.
	Base []string
	// Extra holds KEY=VALUE overrides; applied last so they win over PATH
	// manipulation and over the inherited interactive shell env.
	Extra map[string]string
	// ExtraPath is a PATH-style list of directories prepended to PATH, for
	// tool shims (nvm/volta/bin dirs) this service must see on top of the
	// host's PATH.
	ExtraPath string
	// InheritShellEnv merges the environment of an interactive login shell
	// ($SHELL -lic 'env -0'), which sources ~/.zshrc and ~/.bashrc. A plain
	// `bash -lc` only reads login profiles, so tools exported by rc files stay
	// invisible without this.
	InheritShellEnv bool
	// LoginEnv overrides login-shell discovery. Nil uses lookpath defaults.
	LoginEnv *lookpath.LoginEnvOptions
	// LookupTools overrides DefaultManagedCLIs. Nil uses the default list; an
	// empty non-nil slice disables the lookup step.
	LookupTools []string
}

// BuildEnv returns an environment suitable for `bash -lc` managed jobs:
//
//  1. Base (or os.Environ) as handed in
//  2. interactive login shell env (opt-in), later keys win
//  3. Extra overrides
//  4. ExtraPath dirs prepended to PATH (cleaned, deduped, ~ expanded)
//  5. dirs for LookupTools that are still missing from PATH
//
// The returned error reports only an interactive-shell probe failure. The env
// is usable regardless: every other layer applies, and bash -lc still reads
// login profiles on its own.
func BuildEnv(opts EnvOptions) ([]string, error) {
	base := opts.Base
	if base == nil {
		base = os.Environ()
	}
	env := append([]string(nil), base...)

	var shellErr error
	if opts.InheritShellEnv {
		shellEnv, err := resolveLoginEnv(opts.LoginEnv)
		if err != nil {
			shellErr = err
		} else {
			env = lookpath.MergeEnvs(env, shellEnv)
		}
	}

	env = mergeExtraEnv(env, opts.Extra)

	if extraPath := strings.TrimSpace(opts.ExtraPath); extraPath != "" {
		dirs, err := extraPathDirs(extraPath)
		if err != nil {
			// Callers validate before saving; a stale record must not break the
			// process env, so skip only the bad layer.
			fmt.Fprintf(os.Stderr, "[service] ignoring extra path %q: %v\n", extraPath, err)
		} else {
			env = prependPathDirs(env, dirs)
		}
	}

	tools := opts.LookupTools
	if tools == nil {
		tools = DefaultManagedCLIs
	}
	env = enrichWithLookupPaths(env, tools)
	return env, shellErr
}

func resolveLoginEnv(opts *lookpath.LoginEnvOptions) ([]string, error) {
	var loginOpts lookpath.LoginEnvOptions
	if opts != nil {
		loginOpts = *opts
	}
	_, env, err := lookpath.ResolveLoginEnvs(loginOpts)
	if err != nil {
		return nil, err
	}
	if len(env) == 0 {
		return nil, fmt.Errorf("interactive shell reported an empty environment")
	}
	return env, nil
}

// NormalizeExtraPath cleans a PATH-style directory list: it trims spaces,
// expands a leading ~, resolves ., drops duplicates and returns the remaining
// directories joined by the platform list separator.
func NormalizeExtraPath(extraPath string) (string, error) {
	dirs, err := extraPathDirs(extraPath)
	if err != nil {
		return "", err
	}
	return strings.Join(dirs, string(os.PathListSeparator)), nil
}

// extraPathDirs validates and normalizes the segments of a PATH-style list.
func extraPathDirs(extraPath string) ([]string, error) {
	trimmed := strings.TrimSpace(extraPath)
	if trimmed == "" {
		return nil, nil
	}
	if strings.ContainsAny(trimmed, "\n\r") {
		return nil, fmt.Errorf("extra path must be a single line")
	}
	raw := strings.Split(trimmed, string(os.PathListSeparator))
	dirs := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, segment := range raw {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return nil, fmt.Errorf("extra path has an empty entry; separate directories with %q", string(os.PathListSeparator))
		}
		if strings.ContainsAny(segment, "$`") {
			return nil, fmt.Errorf("extra path does not expand variables: %s", segment)
		}
		expanded, err := expandHome(segment)
		if err != nil {
			return nil, err
		}
		cleaned := filepath.Clean(expanded)
		if cleaned == "/" {
			return nil, fmt.Errorf("invalid extra path directory: %s", segment)
		}
		if !filepath.IsAbs(cleaned) {
			// A relative entry would resolve against the host's cwd, not the
			// service working dir, so refuse it instead of guessing.
			return nil, fmt.Errorf("extra path directory must be absolute: %s", segment)
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		dirs = append(dirs, cleaned)
	}
	return dirs, nil
}

// expandHome replaces a leading "~" with the user's home directory. A bare "~"
// resolves to home itself.
func expandHome(dir string) (string, error) {
	if dir != "~" && !strings.HasPrefix(dir, "~/") {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("cannot resolve ~ in extra path %q: %v", dir, err)
	}
	if dir == "~" {
		return home, nil
	}
	return filepath.Join(home, dir[2:]), nil
}

func mergeExtraEnv(env []string, extra map[string]string) []string {
	if len(extra) == 0 {
		return env
	}
	type entry struct{ key, value string }
	entries := make([]entry, 0, len(extra))
	for key, value := range extra {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		entries = append(entries, entry{key: key, value: value})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })
	overrides := make([]string, 0, len(entries))
	for _, item := range entries {
		overrides = append(overrides, item.key+"="+item.value)
	}
	// MergeEnvs keeps first-seen key order and lets later slices win, so the
	// existing PATH stays in place while an extra PATH entry replaces it.
	return lookpath.MergeEnvs(env, overrides)
}

func enrichWithLookupPaths(env []string, names []string) []string {
	if len(names) == 0 {
		return env
	}
	// Only probe names not already found on the current PATH.
	need := namesMissingOnPath(env, names)
	if len(need) == 0 {
		return env
	}
	items, err := lookpath.LookupPaths(need, lookpath.Options{})
	if err != nil || len(items) == 0 {
		return env
	}
	dirs := items.Dirs()
	if len(dirs) == 0 {
		return env
	}
	return prependPathDirs(env, dirs)
}

func namesMissingOnPath(env []string, names []string) []string {
	pathVal := EnvValue(env, "PATH")
	var need []string
	for _, name := range names {
		if name == "" {
			continue
		}
		if findExecutableOnPath(pathVal, name) {
			continue
		}
		need = append(need, name)
	}
	return need
}

func findExecutableOnPath(pathVal, name string) bool {
	if pathVal == "" {
		return false
	}
	for _, dir := range filepath.SplitList(pathVal) {
		if dir == "" {
			continue
		}
		if lookpath.IsExecutable(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

func prependPathDirs(env []string, dirs []string) []string {
	if len(dirs) == 0 {
		return env
	}
	pathVal := EnvValue(env, "PATH")
	existing := make(map[string]struct{})
	for _, d := range filepath.SplitList(pathVal) {
		if d == "" {
			continue
		}
		existing[filepath.Clean(d)] = struct{}{}
	}
	var prefix []string
	for _, d := range dirs {
		d = filepath.Clean(d)
		if d == "" || d == "." {
			continue
		}
		if _, ok := existing[d]; ok {
			continue
		}
		existing[d] = struct{}{}
		prefix = append(prefix, d)
	}
	if len(prefix) == 0 {
		return env
	}
	newPath := strings.Join(prefix, string(os.PathListSeparator))
	if pathVal != "" {
		newPath = newPath + string(os.PathListSeparator) + pathVal
	}
	return SetEnvValue(env, "PATH", newPath)
}
