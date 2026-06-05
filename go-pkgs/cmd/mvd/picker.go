package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	llsrun "github.com/xhd2015/lls/run"
	"golang.org/x/term"
)

type pickerEntry struct {
	display string
	full    string
}

func buildProjectList(hist History, aliases map[string]string) []pickerEntry {
	aliasesByRoot := groupAliasesByRoot(aliases)

	entries := make([]pickerEntry, 0, len(hist))
	seen := make(map[string]bool)

	for _, locations := range hist {
		if len(locations) == 0 {
			continue
		}
		root := locations[0].Path
		latest := locations[len(locations)-1].Path
		disp := displayPath(latest)
		if seen[disp] {
			continue
		}
		seen[disp] = true

		if aliasList := aliasesByRoot[root]; len(aliasList) > 0 {
			disp = fmt.Sprintf("%s (aliases: %s)", disp, joinAliases(aliasList))
		}

		entries = append(entries, pickerEntry{display: disp, full: latest})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].display < entries[j].display
	})

	return entries
}

func joinAliases(aliases []string) string {
	result := ""
	for i, a := range aliases {
		if i > 0 {
			result += ", "
		}
		result += a
	}
	return result
}

type pickAction func(fullPath string) error

func runPicker(action pickAction) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		return fmt.Errorf("no tracked projects, use `mvd --add DIR` to add one")
	}

	aliases, err := loadAliases()
	if err != nil {
		return err
	}

	entries := buildProjectList(hist, aliases)

	options := make([]string, len(entries))
	for i, e := range entries {
		options[i] = e.display
	}

	selected, err := llsrun.SelectWithFzf(options, "")
	if err != nil {
		return err
	}
	if selected == "" {
		return nil
	}

	for _, e := range entries {
		if e.display == selected {
			return action(e.full)
		}
	}

	return fmt.Errorf("selected project not found: %s", selected)
}

func cmdPickAndPrint() error {
	if !stdinIsTerminal() {
		return fmt.Errorf("usage: mvd --print SRC")
	}
	return runPicker(func(fullPath string) error {
		fmt.Printf("%s -> %s\n", displayPath(fullPath), fullPath)
		return nil
	})
}

func cmdPickAndVscode() error {
	if !stdinIsTerminal() {
		return fmt.Errorf("usage: mvd --vscode SRC")
	}
	return runPicker(func(fullPath string) error {
		fmt.Printf("%s -> %s\n", displayPath(fullPath), fullPath)
		cmd := exec.Command("code", fullPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("open vscode: %w", err)
		}
		return nil
	})
}

func cmdPickAndCd() error {
	if !stdinIsTerminal() {
		return fmt.Errorf("usage: mvd --cd SRC")
	}
	return runPicker(func(fullPath string) error {
		fmt.Printf("%s -> %s\n", displayPath(fullPath), fullPath)
		return launchShell(fullPath, displayPath(fullPath))
	})
}

func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
