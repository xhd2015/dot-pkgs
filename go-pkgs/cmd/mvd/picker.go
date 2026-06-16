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

		hasWorktree := false
		mainRepos := make(map[string]bool)
		for _, loc := range locations {
			if loc.Git != nil && loc.Git.Type == "worktree" {
				hasWorktree = true
				mainRepos[loc.Git.MainRepo] = true
			}
		}

		var paths []string
		if hasWorktree {
			paths = append(paths, root)
			for _, loc := range locations {
				if loc.Git != nil && loc.Git.Type == "worktree" {
					paths = append(paths, loc.Path)
				}
			}
			for _, loc := range locations {
				if loc.Path == root {
					continue
				}
				if loc.Git != nil && loc.Git.Type == "worktree" {
					continue
				}
				if mainRepos[loc.Path] {
					paths = append(paths, loc.Path)
				}
			}
			if !containsPathStr(paths, latest) {
				paths = append(paths, latest)
			}
		} else {
			paths = append(paths, latest)
		}

		aliasList := aliasesByRoot[root]

		for _, path := range paths {
			disp := displayPath(path)
			if seen[disp] {
				continue
			}
			seen[disp] = true

			loc := findLocation(locations, path)
			marker := computePickerMarker(path, loc, hasWorktree, root, aliasList)
			if marker != "" {
				disp = disp + " " + marker
			}

			entries = append(entries, pickerEntry{display: disp, full: path})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].display < entries[j].display
	})

	return entries
}

func pathDead(p string) bool {
	_, err := os.Stat(p)
	return err != nil
}

func findLocation(locations []LocationEntry, path string) *LocationEntry {
	for i := range locations {
		if locations[i].Path == path {
			return &locations[i]
		}
	}
	return nil
}

func computePickerMarker(path string, loc *LocationEntry, hasWorktree bool, root string, aliasList []string) string {
	isDead := pathDead(path)
	isWt := loc != nil && loc.Git != nil && loc.Git.Type == "worktree"
	isMain := hasWorktree && !isWt
	isRoot := path == root
	isExtMain := isMain && !isRoot
	hasAliases := len(aliasList) > 0 && (path == root || !hasWorktree)

	aliasSuffix := ""
	if hasAliases {
		aliasSuffix = ", aliases: " + joinAliases(aliasList)
	}

	if isDead && isWt {
		if hasAliases {
			return "(dead worktree" + aliasSuffix + ")"
		}
		return "(dead worktree)"
	}
	if isDead && isRoot && isMain {
		if hasAliases {
			return "(dead main" + aliasSuffix + ")"
		}
		return "(dead main)"
	}
	if isDead && isExtMain {
		if hasAliases {
			return "(dead external main" + aliasSuffix + ")"
		}
		return "(dead external main)"
	}
	if isDead {
		if hasAliases {
			return "(dead" + aliasSuffix + ")"
		}
		return "(dead)"
	}
	if isWt {
		if hasAliases {
			return "(worktree" + aliasSuffix + ")"
		}
		return "(worktree)"
	}
	if isRoot && isMain {
		if hasAliases {
			return "(main" + aliasSuffix + ")"
		}
		return "(main)"
	}
	if isExtMain {
		if hasAliases {
			return "(external main" + aliasSuffix + ")"
		}
		return "(external main)"
	}
	if hasAliases {
		return "(aliases: " + joinAliases(aliasList) + ")"
	}
	return ""
}

func containsPathStr(paths []string, target string) bool {
	for _, p := range paths {
		if p == target {
			return true
		}
	}
	return false
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
	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		return fmt.Errorf("no tracked projects, use `mvd --add DIR` to add one")
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
