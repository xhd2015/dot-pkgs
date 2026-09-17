package main

import (
	"fmt"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/vscode"
)

func cmdVscode(src string) error {
	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}

	_, _, lastLoc, err := resolveMoveSource(hist, aliases, src)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("dry-run: would open VSCode at %s\n", displayPath(lastLoc))
		return nil
	}

	// Wait for the CLI so a failed launch is reported, not swallowed.
	if _, err := vscode.OpenConfig(lastLoc, &vscode.Config{Run: vscode.WaitRunner}); err != nil {
		return fmt.Errorf("open vscode: %w", err)
	}
	return nil
}
