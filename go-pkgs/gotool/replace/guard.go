package replace

import (
	"os"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

// ReplaceIssue describes a local filesystem replace directive found in a
// go.mod file.  IsIntraRepo is true when the replace target resolves to an
// existing directory that shares the same git toplevel as the scan root.
type ReplaceIssue struct {
	GoModPath   string // absolute path to the go.mod file
	OldPath     string // replace old path
	NewPath     string // replace new path (the local target)
	IsIntraRepo bool   // true if target is inside the same git repo
}

// CheckLocalReplaces scans all Go modules under top for local filesystem
// replace directives and returns every issue found.  The caller decides
// policy (lenient, strict, etc.).
func CheckLocalReplaces(top string) ([]ReplaceIssue, error) {
	absTop, err := filepath.Abs(top)
	if err != nil {
		return nil, err
	}

	// Resolve the repo's git toplevel once, canonicalizing symlinks so
	// comparisons work on platforms where temp dirs are symlinked (e.g.
	// macOS /var -> /private/var).
	repoTop, err := worktree.ShowToplevel(absTop)
	if err != nil {
		return nil, err
	}

	modules, err := scan.Scan(absTop, scan.Options{})
	if err != nil {
		return nil, err
	}

	var issues []ReplaceIssue
	for _, m := range modules {
		offenders := m.LocalFilesystemReplaces()
		if len(offenders) == 0 {
			continue
		}

		modDir := filepath.Join(absTop, m.Dir)
		goModPath := filepath.Join(modDir, "go.mod")

		for _, r := range offenders {
			issues = append(issues, ReplaceIssue{
				GoModPath:   goModPath,
				OldPath:     r.OldPath,
				NewPath:     r.NewPath,
				IsIntraRepo: isIntraRepoReplace(modDir, r.NewPath, repoTop),
			})
		}
	}
	return issues, nil
}

// isIntraRepoReplace reports whether the replace target at replacePath
// (relative to modDir, or absolute) resolves to an existing directory that
// lives in the same git repo as consumerTop.
func isIntraRepoReplace(modDir, replacePath, consumerTop string) bool {
	target := replacePath
	if !filepath.IsAbs(target) {
		target = filepath.Join(modDir, target)
	}
	target = filepath.Clean(target)
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return false
	}
	top, err := worktree.ShowToplevel(target)
	if err != nil {
		return false
	}
	return top == consumerTop
}
