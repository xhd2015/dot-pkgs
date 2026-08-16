package pathfmt

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// envNameOK matches shell-style identifiers used as display aliases.
var envNameOK = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// skipExactEnvNames are never used as $NAME path aliases.
var skipExactEnvNames = map[string]struct{}{
	"HOME": {}, "PWD": {}, "OLDPWD": {}, "PATH": {}, "CDPATH": {},
	"FPATH": {}, "MANPATH": {}, "INFOPATH": {}, "TMPDIR": {}, "TMP": {},
	"TEMP": {}, "SHELL": {}, "TERM": {}, "USER": {}, "LOGNAME": {},
	"_": {}, "SHLVL": {},
}

// secretEnvSubstrings mark secret-ish names (case-sensitive substring).
var secretEnvSubstrings = []string{"KEY", "TOKEN", "SECRET", "PASSWORD", "AUTH"}

func evalPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	// Missing leaf paths (e.g. not-yet-installed integration targets) may not
	// resolve via EvalSymlinks; canonicalize through the longest existing prefix.
	parts := strings.Split(path, string(filepath.Separator))
	for i := len(parts); i > 0; i-- {
		prefix := strings.Join(parts[:i], string(filepath.Separator))
		if prefix == "" {
			prefix = string(filepath.Separator)
		}
		resolved, err := filepath.EvalSymlinks(prefix)
		if err != nil {
			continue
		}
		suffix := strings.Join(parts[i:], string(filepath.Separator))
		if suffix == "" {
			return resolved
		}
		return filepath.Join(resolved, suffix)
	}
	return path
}

// Short returns a shortened form of path for human-readable CLI output only.
// It must not be used for file I/O or exec.Command.Dir.
// Equivalent to ShortFrom(path, "") — base is process cwd.
func Short(path string) string {
	return ShortFrom(path, "")
}

// ShortFrom shortens path for display relative to baseDir, then home.
// It must not be used for file I/O or exec.Command.Dir.
//
// Rules:
//  1. Normalize path with filepath.Abs.
//  2. base = baseDir when non-empty, else os.Getwd().
//  3. If base is usable and is not the user home directory, try cwd-style
//     relative: path == base → "."; strict child → rel (no ".." prefix).
//     When base equals home, skip relative form so under-home paths become
//     "~/..." instead of ".spl/..." (cross-process / agent prompts).
//  4. If under home → "~" or "~/...".
//  5. Else absolute path.
func ShortFrom(path, baseDir string) string {
	if path == "" {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	base := strings.TrimSpace(baseDir)
	if base == "" {
		if cwd, err := os.Getwd(); err == nil {
			base = cwd
		}
	}
	if base != "" {
		if baseAbs, err := filepath.Abs(base); err == nil {
			baseEval := evalPath(baseAbs)
			absEval := evalPath(abs)
			if !baseIsHome(baseEval, baseAbs) {
				if absEval == baseEval {
					return "."
				}
				rel, err := filepath.Rel(baseEval, absEval)
				if err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
					return rel
				}
			}
		}
	}

	return tildeHomeAbs(abs)
}

// TildeHome shortens path for human-readable display by replacing the
// user home directory prefix with "~". It does not produce cwd-relative
// forms. Display-only — do not use for file I/O or exec.Command.Dir.
func TildeHome(path string) string {
	if path == "" {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return tildeHomeAbs(abs)
}

// tildeHomeAbs applies home-tilde rules to an already-absolute path.
// Returns "~", "~/...", or abs. Shared by TildeHome and ShortFrom.
func tildeHomeAbs(abs string) string {
	if home, err := os.UserHomeDir(); err == nil {
		if abs == home {
			return "~"
		}
		homePrefix := home + string(filepath.Separator)
		if strings.HasPrefix(abs, homePrefix) {
			return "~" + strings.TrimPrefix(abs, home)
		}
	}
	return abs
}

// baseIsHome reports whether base (already Abs/eval candidates) is the user home.
func baseIsHome(baseEval, baseAbs string) bool {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return false
	}
	homeAbs, err := filepath.Abs(home)
	if err != nil {
		return false
	}
	homeEval := evalPath(homeAbs)
	return baseEval == homeEval || baseAbs == homeAbs || baseAbs == home || baseEval == home
}

// Expand converts a display path (with ~ prefix) back to an absolute path
// for filesystem operations. Non-display paths are returned unchanged.
func Expand(path string) string {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	suffix := strings.TrimPrefix(path, "~")
	suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
	if suffix == "" {
		return home
	}
	return filepath.Join(home, suffix)
}

// ShortEnv shortens path for display using eligible path aliases from
// os.Environ(). Display-only — do not use for file I/O or exec.Command.Dir.
// Equivalent to ShortEnvFrom(path, os.Environ()).
func ShortEnv(path string) string {
	return ShortEnvFrom(path, os.Environ())
}

// ShortEnvFrom shortens path for display using eligible path aliases from env
// (KEY=value pairs, os.Environ style). Longest segment-boundary prefix match
// becomes $NAME + remainder; otherwise TildeHome(path). Empty/nil env does not
// fall back to os.Environ. Not cwd-relative. Display-only.
func ShortEnvFrom(path string, env []string) string {
	if path == "" {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	homeAbs := userHomeAbs()
	bestName, bestPrefix := "", ""
	bestLen := -1
	for _, entry := range env {
		name, value, ok := parseEnvEntry(entry)
		if !ok || !envAliasEligible(name, value, homeAbs) {
			continue
		}
		if !isSegmentPrefix(abs, value) {
			continue
		}
		n := len(value)
		if n > bestLen ||
			(n == bestLen && (bestName == "" ||
				len(name) < len(bestName) ||
				(len(name) == len(bestName) && name < bestName))) {
			bestLen = n
			bestName = name
			bestPrefix = value
		}
	}
	if bestName == "" {
		return tildeHomeAbs(abs)
	}
	if abs == bestPrefix {
		return "$" + bestName
	}
	return "$" + bestName + strings.TrimPrefix(abs, bestPrefix)
}

func parseEnvEntry(entry string) (name, value string, ok bool) {
	i := strings.IndexByte(entry, '=')
	if i <= 0 {
		return "", "", false
	}
	return entry[:i], entry[i+1:], true
}

func envAliasEligible(name, value, homeAbs string) bool {
	if !envNameOK.MatchString(name) {
		return false
	}
	if _, skip := skipExactEnvNames[name]; skip {
		return false
	}
	for _, sub := range secretEnvSubstrings {
		if strings.Contains(name, sub) {
			return false
		}
	}
	if value == "" || !filepath.IsAbs(value) {
		return false
	}
	if strings.Contains(value, ":") || strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		return false
	}
	if homeAbs != "" && valueEqualsHome(value, homeAbs) {
		return false
	}
	return true
}

func userHomeAbs() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	if abs, err := filepath.Abs(home); err == nil {
		return abs
	}
	return home
}

func valueEqualsHome(value, homeAbs string) bool {
	if value == homeAbs {
		return true
	}
	if abs, err := filepath.Abs(value); err == nil && abs == homeAbs {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil && (value == home || value == homeAbs) {
		return true
	}
	return false
}

// isSegmentPrefix reports whether abs equals prefix or is a child under it.
func isSegmentPrefix(abs, prefix string) bool {
	if abs == prefix {
		return true
	}
	return strings.HasPrefix(abs, prefix+string(filepath.Separator))
}
