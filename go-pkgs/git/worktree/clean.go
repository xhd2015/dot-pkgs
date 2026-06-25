package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsClean reports whether the worktree has no uncommitted changes.
func IsClean(path string) error {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return fmt.Errorf("worktree %s has uncommitted changes", path)
	}
	return nil
}