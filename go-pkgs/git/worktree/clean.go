package worktree

import (
	"context"
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

// IsClean reports whether the worktree has no uncommitted changes.
func IsClean(path string) error {
	out, err := cmd.Run(context.Background(), path, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(out)) > 0 {
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
	return counts.Added == 0 && counts.Changed == 0 && counts.Renamed == 0 && counts.Deleted == 0, nil
}