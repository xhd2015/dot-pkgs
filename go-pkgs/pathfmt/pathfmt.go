package pathfmt

import (
	"os"
	"path/filepath"
	"strings"
)

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
