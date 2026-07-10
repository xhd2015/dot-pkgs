package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/interactive"
)

func cmdCd(src string) error {
	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}

	_, _, lastLoc, err := resolveMoveSource(hist, aliases, src)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("dry-run: would cd to %s\n", displayPath(lastLoc))
		return nil
	}

	return launchShell(lastLoc, src)
}

func launchShell(dir string, name string) error {
	err := interactive.LoginInteractive(dir, name, "MVD_SHELL=1", "MVD_PROJECT="+name)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("cd: %w", err)
	}
	return nil
}
