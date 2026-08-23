package worktree

import (
	"context"
	"fmt"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
	gitopsWorktree "github.com/xhd2015/gitops/git/worktree"
)

// IsClean reports whether the worktree has no uncommitted changes.
// Uses porcelain status (untracked counts as dirty), matching historical go-pkgs
// semantics via gitops IsPorcelainClean.
func IsClean(path string) error {
	ok, err := gitopsWorktree.IsPorcelainClean(path)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if !ok {
		return fmt.Errorf("worktree %s has uncommitted changes", path)
	}
	return nil
}

// IsCleanWrk reports whether the worktree is clean under wrk status taxonomy.
func IsCleanWrk(path string) (bool, error) {
	out, err := cmd.Run(context.Background(), path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	counts := status.ParsePorcelainWrk(out)
	return counts.Staged == 0 && counts.Changed == 0 && counts.Renamed == 0 && counts.Deleted == 0 && counts.Untracked == 0, nil
}
