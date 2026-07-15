package tagscope

import (
	"fmt"
	"os/exec"
	"strings"
)

// Collect runs `git tag -l` in repoRoot and delegates to CollectFromNames.
func Collect(repoRoot string) (CollectedTags, error) {
	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return CollectedTags{}, fmt.Errorf("git tag -l in %s: %w", repoRoot, err)
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return CollectFromNames(nil), nil
	}

	names := strings.Split(text, "\n")
	return CollectFromNames(names), nil
}