package wrkcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/interactive"
)

// runCd jumps into absDir: in-place follow-up when WRK_FOLLOWUP_FILE is set,
// otherwise prints install hint + abs path and launches an interactive shell.
func runCd(absDir string) error {
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wrk: %s does not exist", absDir)
		}
		return fmt.Errorf("stat dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("wrk: %s is not a directory", absDir)
	}

	// Branch A — bash integration channel open: in-place follow-up only.
	if strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE")) != "" {
		return writeFollowupCD(false, absDir)
	}

	// Branch B — fallback: warn, print abs path, launch interactive shell.
	fmt.Fprintf(os.Stderr, "warning: bash integration not active; install with: wrk --bash-integration --install\n")
	fmt.Fprintln(os.Stdout, absDir)

	err = interactive.LoginInteractive(absDir, filepath.Base(absDir), "WRK_SHELL=1")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}
