package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdMove(src, dst string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	aliases, err := loadAliases()
	if err != nil {
		return err
	}

	origKey, locations, absSrc, err := resolveMoveSource(hist, aliases, src)
	if err != nil {
		return err
	}

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve dst: %w", err)
	}

	info, err := os.Stat(absDst)
	if err == nil && info.IsDir() {
		absDst = filepath.Join(absDst, filepath.Base(absSrc))
	}

	isWt := isGitWorktree(absSrc)
	if err := moveDir(absSrc, absDst); err != nil {
		return err
	}
	if isWt {
		fmt.Printf("moved worktree: %s → %s\n", displayPath(absSrc), displayPath(absDst))
	} else {
		fmt.Printf("moved: %s → %s\n", displayPath(absSrc), displayPath(absDst))
	}

	if locations == nil {
		origKey = absSrc
		locations = []LocationEntry{{Path: absSrc}, {Path: absDst}}
	} else {
		locations = append(locations, LocationEntry{Path: absDst})
	}

	delete(hist, origKey)
	hist[locations[0].Path] = locations

	return saveHistory(hist)
}

func cmdBack(src string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations, err := resolveBackEntry(hist, src)
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		return fmt.Errorf("empty mv history for %s", src)
	}
	last := locations[len(locations)-1]
	if len(locations) <= 1 {
		fmt.Printf("nothing to move back for %s\n", displayPath(last.Path))
		return nil
	}

	prev := locations[len(locations)-2]

	if last.Git != nil && last.Git.Type == "worktree" {
		return cmdWorktreeBack(origKey, locations)
	}

	isWt := isGitWorktree(last.Path)
	if err := moveDir(last.Path, prev.Path); err != nil {
		return err
	}
	if isWt {
		fmt.Printf("moved worktree back: %s → %s\n", displayPath(last.Path), displayPath(prev.Path))
	} else {
		fmt.Printf("moved back: %s → %s\n", displayPath(last.Path), displayPath(prev.Path))
	}

	locations = locations[:len(locations)-1]
	hist[origKey] = locations

	return saveHistory(hist)
}

func moveDir(src, dst string) error {
	if isGitWorktree(src) {
		return moveWorktree(src, dst)
	}

	var wts []worktreeInfo
	if isGitRepo(src) {
		var err error
		wts, err = listWorktrees(src)
		if err != nil {
			wts = nil
		}
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("move: %w", err)
	}

	for _, wt := range wts {
		if err := updateWorktreeGitFile(wt.path, dst); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to update worktree %s: %v\n", wt.path, err)
			continue
		}
		fmt.Printf("  updated worktree: %s\n", displayPath(wt.path))
	}

	return nil
}

func moveWorktree(src, dst string) error {
	mainRepo, err := readWorktreeMainRepo(src)
	if err != nil {
		return fmt.Errorf("read worktree main repo: %w", err)
	}
	cmd := exec.Command("git", "-C", mainRepo, "worktree", "move", src, dst)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree move: %w\n%s", err, out)
	}
	return nil
}
