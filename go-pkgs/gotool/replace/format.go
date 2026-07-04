package replace

import (
	"path/filepath"
	"strings"
)

func canonicalPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolved)
	}
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

// GoModRelPath returns the go.mod path relative to scanTop for display.
// The root module is "go.mod"; nested modules use slash-separated paths such as
// "sub/go.mod".
func GoModRelPath(scanTop, goModPath string) string {
	scanTop = canonicalPath(scanTop)
	goModPath = canonicalPath(goModPath)
	rel, err := filepath.Rel(scanTop, goModPath)
	if err != nil {
		return filepath.ToSlash(goModPath)
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return filepath.ToSlash(goModPath)
	}
	return rel
}

// ResolvedTargetDir returns the absolute resolved directory targeted by issue.
func ResolvedTargetDir(issue ReplaceIssue) string {
	modDir := filepath.Dir(issue.GoModPath)
	target := issue.NewPath
	if !filepath.IsAbs(target) {
		target = filepath.Join(modDir, target)
	}
	target = filepath.Clean(target)
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		return resolved
	}
	if abs, err := filepath.Abs(target); err == nil {
		return abs
	}
	return target
}

// FormatIssueLine formats a replace issue for human-facing output:
//
//	go.mod: => /abs/target
//	sub/go.mod: => /abs/target
func FormatIssueLine(scanTop string, issue ReplaceIssue) string {
	return GoModRelPath(scanTop, issue.GoModPath) + ": => " + ResolvedTargetDir(issue)
}