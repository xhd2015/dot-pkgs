package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeFollowupCD appends a single "cd /absolute/path" line to WRK_FOLLOWUP_FILE
// when the channel is set and follow-ups are not disabled via --no-cd.
// No-op when the env is unset/empty or disabled is true.
func writeFollowupCD(disabled bool, absPath string) error {
	if disabled {
		return nil
	}
	outPath := strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE"))
	if outPath == "" {
		return nil
	}
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return nil
	}
	if !filepath.IsAbs(absPath) {
		resolved, err := filepath.Abs(absPath)
		if err != nil {
			return fmt.Errorf("resolve follow-up path: %w", err)
		}
		absPath = resolved
	}
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open follow-up file: %w", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "cd %s\n", absPath); err != nil {
		return fmt.Errorf("write follow-up: %w", err)
	}
	return nil
}

// shouldWriteCwdGatedFollowup reports whether a done/set-task follow-up cd should
// be written: true only when shellCwdAtStart is non-empty and no longer exists
// on the filesystem (os.Stat not-exist). Empty path or still-existing path → false.
func shouldWriteCwdGatedFollowup(shellCwdAtStart string) bool {
	shellCwdAtStart = strings.TrimSpace(shellCwdAtStart)
	if shellCwdAtStart == "" {
		return false
	}
	_, err := os.Stat(shellCwdAtStart)
	return os.IsNotExist(err)
}

// writeFollowupCDIfCwdMissing writes cd dest only when shellCwd no longer exists.
// Used after successful --done remove / --set-task move so a surviving sibling
// or main checkout is not yanked by auto-cd. Create paths must use writeFollowupCD.
func writeFollowupCDIfCwdMissing(disabled bool, shellCwd, dest string) error {
	if !shouldWriteCwdGatedFollowup(shellCwd) {
		return nil
	}
	return writeFollowupCD(disabled, dest)
}
