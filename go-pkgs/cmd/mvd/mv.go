package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func cmdMove(src, dst string) error {
	hist, aliases, err := loadHistory()
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

	// Read worktree Git metadata before moving (for locations == nil case).
	var srcGitInfo *GitInfo
	if isWt && locations == nil {
		gitInfo, err := readWorktreeGitInfo(absSrc)
		if err != nil {
			return fmt.Errorf("read worktree git info for %s: %w", displayPath(absSrc), err)
		}
		srcGitInfo = gitInfo
	}

	if dryRun {
		fmt.Printf("dry-run: would move %s -> %s\n", displayPath(absSrc), displayPath(absDst))
		return nil
	}

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
		if isWt {
			locations = []LocationEntry{{Path: absSrc, Git: srcGitInfo}, {Path: absDst, Git: srcGitInfo}}
		} else {
			locations = []LocationEntry{{Path: absSrc}, {Path: absDst}}
		}
	} else {
		var dstGitInfo *GitInfo
		if isWt {
			for i := len(locations) - 1; i >= 0; i-- {
				if locations[i].Path == absSrc && locations[i].Git != nil && locations[i].Git.Type == "worktree" {
					dstGitInfo = locations[i].Git
					break
				}
			}
		}
		locations = append(locations, LocationEntry{Path: absDst, Git: dstGitInfo})
	}

	delete(hist, origKey)
	hist[locations[0].Path] = locations

	return saveHistory(hist, aliases)
}

func cmdBack(src string) error {
	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations, absSrc, err := resolveBackEntry(hist, aliases, src)
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

	// Check if the user targeted a worktree entry (may be last or non-last).
	// Worktrees are independent branches that can be removed at any position.
	var wtIdx int = -1
	for i, loc := range locations {
		if loc.Path == absSrc && loc.Git != nil && loc.Git.Type == "worktree" {
			wtIdx = i
			break
		}
	}
	if wtIdx >= 0 && wtIdx == len(locations)-1 {
		return cmdWorktreeBack(origKey, locations)
	}
	if wtIdx >= 0 {
		// Non-last worktree: remove it from the chain while keeping
		// subsequent moves intact.
		wtLoc := locations[wtIdx]
		return cmdWorktreeBackAt(origKey, locations, wtIdx, wtLoc)
	}

	prev := locations[len(locations)-2]

	for i := len(locations) - 2; i >= 0; i-- {
		loc := locations[i]
		if loc.Git == nil || loc.Git.Type != "worktree" {
			prev = loc
			break
		}
	}

	if dryRun {
		fmt.Printf("dry-run: would move back %s\n", displayPath(last.Path))
		return nil
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

	// If this back restored a parent directory, clean up orphaned sub-project
	// entries that were created as separate entries during the parent move.
	// These sub-projects were physically moved/restored with the parent but
	// never had their own history entries updated — they're now redundant.
	for key := range hist {
		if key != origKey && isUnderDir(prev.Path, key) {
			delete(hist, key)
		}
	}

	return saveHistory(hist, aliases)
}

func moveDir(src, dst string) error {
	// Resolve symlinks to get canonical paths. On macOS /tmp is a symlink
	// to /private/tmp, and git may return paths with a different prefix
	// than Go's filepath.Abs. Normalizing both ensures correct comparisons.
	if resolved, err := filepath.EvalSymlinks(src); err == nil {
		src = resolved
	}
	if resolved, err := filepath.EvalSymlinks(dst); err == nil {
		dst = resolved
	}

	// Collect all worktrees from git repos under src (including src itself).
	// This handles parent-directory moves where the parent itself isn't a git
	// repo but contains subdirectories that are (with linked worktrees).
	type wtRec struct {
		wtPath   string
		mainRepo string
	}
	var wtRecs []wtRec

	addWorktrees := func(repoPath string) {
		wts, err := listWorktrees(repoPath)
		if err != nil {
			return
		}
		for _, wt := range wts {
			// Normalize worktree path (git may return /private/tmp while
			// Go uses /tmp on macOS).
			wtPath := wt.path
			if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
				wtPath = resolved
			}
			// Deduplicate: a worktree can be listed by multiple repos
			seen := false
			for _, r := range wtRecs {
				if r.wtPath == wtPath {
					seen = true
					break
				}
			}
			if !seen {
				wtRecs = append(wtRecs, wtRec{wtPath: wtPath, mainRepo: repoPath})
			}
		}
	}

	if isGitRepo(src) {
		addWorktrees(src)
	}

	// Also check immediate subdirectories for nested git repos.
	// This handles the case where src is a parent directory containing
	// git repos (and their worktrees) but is not itself a git repo.
	if entries, err := os.ReadDir(src); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			subPath := filepath.Join(src, e.Name())
			if isGitRepo(subPath) {
				addWorktrees(subPath)
			}
		}
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("move: %w", err)
	}

	// Update worktree .git files. If the worktree or its main repo were
	// under the moved directory, their paths are now under dst.
	for _, r := range wtRecs {
		newWtPath := r.wtPath
		newMainRepo := r.mainRepo
		if isUnderDir(src, r.wtPath) {
			newWtPath = filepath.Join(dst, r.wtPath[len(src):])
		}
		if isUnderDir(src, r.mainRepo) {
			newMainRepo = filepath.Join(dst, r.mainRepo[len(src):])
		}
		if err := updateWorktreeGitFile(newWtPath, newMainRepo); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to update worktree %s: %v\n", newWtPath, err)
			continue
		}
		fmt.Printf("  updated worktree: %s\n", displayPath(newWtPath))
	}

	return nil
}

// isUnderDir reports whether child equals parent or is a descendant path
// within parent.
func isUnderDir(parent, child string) bool {
	return child == parent || strings.HasPrefix(child, parent+string(filepath.Separator))
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
