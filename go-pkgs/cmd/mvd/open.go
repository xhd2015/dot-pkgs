package main

import (
	"fmt"
	"os"
	"os/exec"
)

func cmdVscode(src string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	aliases, err := loadAliases()
	if err != nil {
		return err
	}

	_, _, lastLoc, err := resolveMoveSource(hist, aliases, src)
	if err != nil {
		return err
	}

	cmd := exec.Command("code", lastLoc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open vscode: %w", err)
	}
	return nil
}
