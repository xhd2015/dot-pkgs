package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	wt "github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/gitops/git"
	"golang.org/x/term"
)

func cmdWorktreeMove(src, dst string) error {
	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}
	if _, locs, ok, err := resolveBasename(hist, src); ok {
		src = locs[len(locs)-1].Path
	} else if err != nil {
		return err
	} else if isBareBaseName(src) {
		if _, err := os.Lstat(src); os.IsNotExist(err) {
			if _, locs, err := findEntryByAlias(hist, aliases, src); err != nil {
				return err
			} else if locs != nil {
				src = locs[len(locs)-1].Path
			}
		}
	}

	srcAbs, err := resolveInputPath(src)
	if err != nil {
		return fmt.Errorf("resolve src: %w", err)
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve dst: %w", err)
	}

	if info, err := os.Stat(dstAbs); err == nil && info.IsDir() {
		dstAbs = filepath.Join(dstAbs, filepath.Base(srcAbs))
	}

	if !isGitRepo(srcAbs) && !isGitWorktree(srcAbs) {
		return fmt.Errorf("%s is not a git repository", displayPath(srcAbs))
	}

	branch, err := generateBranchName(filepath.Base(dstAbs), srcAbs)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("dry-run: would create worktree at %s from %s\n", displayPath(dstAbs), displayPath(srcAbs))
		return nil
	}

	cmd := exec.Command("git", "-C", srcAbs, "worktree", "add", "-b", branch, dstAbs)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, out)
	}

	fmt.Printf("worktree created: %s → %s [branch: %s]\n", displayPath(srcAbs), displayPath(dstAbs), branch)

	entry := LocationEntry{
		Path: dstAbs,
		Git: &GitInfo{
			Type:     "worktree",
			MainRepo: srcAbs,
			Branch:   branch,
		},
	}

	if existing, ok := hist[srcAbs]; ok {
		hist[srcAbs] = append(existing, entry)
	} else {
		hist[srcAbs] = []LocationEntry{{Path: srcAbs}, entry}
	}

	return saveHistory(hist, aliases)
}

func cmdWorktreeBack(origKey string, locations []LocationEntry) error {
	last := locations[len(locations)-1]

	// Read the current main repo from the worktree's .git file on disk.
	// The history's MainRepo may be stale if the main repo was moved
	// via a plain move after the worktree was created.
	mainRepo, err := readWorktreeMainRepo(last.Path)
	if err != nil {
		// Fall back to history value with a warning
		mainRepo = last.Git.MainRepo
		fmt.Fprintf(os.Stderr, "warning: could not read worktree main repo from disk, using history: %s\n", displayPath(mainRepo))
	}

	// Verify the main repo still exists on disk
	if _, err := os.Stat(mainRepo); err != nil {
		return fmt.Errorf("main repo %s no longer exists (worktree references stale path; the main repo may have been moved or deleted)", displayPath(mainRepo))
	}

	branch := last.Git.Branch

	if err := wt.IsClean(last.Path); err != nil {
		return err
	}

	// Determine branch relationship with main HEAD
	result, err := git.CompareBranches(mainRepo, branch, "HEAD")
	if err != nil {
		return err
	}

	switch result.Relation {
	case git.BranchRelationAIsAncestorOfB, git.BranchRelationSame:
		// CASE A: branch is ancestor of HEAD → already merged (existing behavior)
	case git.BranchRelationBIsAncestorOfA:
		// CASE B: HEAD is ancestor of branch → fast-forward possible
		confirmed, err := promptConfirm(fmt.Sprintf("branch %s is ahead, merge and remove?", branch))
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
		cmd := exec.Command("git", "-C", mainRepo, "merge", "--ff-only", branch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git merge --ff-only %s: %w\n%s", branch, err, out)
		}
	case git.BranchRelationDiverged:
		// CASE C: diverged - prompt and try rebase on worktree
		confirmed, err := promptConfirm(fmt.Sprintf("branch %s has diverged, rebase and merge?", branch))
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
		mainCommit, err := revParseCommit(mainRepo, "HEAD")
		if err != nil {
			return err
		}
		rebaseCmd := exec.Command("git", "-C", last.Path, "rebase", mainCommit)
		rebaseOut, rebaseErr := rebaseCmd.CombinedOutput()
		if rebaseErr != nil {
			abortCmd := exec.Command("git", "-C", last.Path, "rebase", "--abort")
			abortCmd.Run()
			return fmt.Errorf("rebase conflict: %w\n%s", rebaseErr, rebaseOut)
		}
		cmd := exec.Command("git", "-C", mainRepo, "merge", "--ff-only", branch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git merge --ff-only %s after rebase: %w\n%s", branch, err, out)
		}
	}

	if dryRun {
		fmt.Printf("dry-run: would remove worktree %s\n", displayPath(last.Path))
		return nil
	}

	cmd := exec.Command("git", "-C", mainRepo, "worktree", "remove", last.Path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove: %w\n%s", err, out)
	}

	cmd = exec.Command("git", "-C", mainRepo, "branch", "-D", branch)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch -D %s: %w\n%s", branch, err, out)
	}

	fmt.Printf("worktree removed: %s [branch: %s deleted]\n", displayPath(last.Path), branch)

	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}

	hist[origKey] = locations[:len(locations)-1]
	if len(hist[origKey]) <= 1 {
		delete(hist, origKey)
	}

	return saveHistory(hist, aliases)
}

// cmdWorktreeBackAt removes a worktree entry at the given index from the chain.
// Unlike cmdWorktreeBack (which always removes the last entry), this handles
// worktrees at any position — e.g. when the chain is [repo, wt(wt), mid] and
// the user does --back on wt in the middle.
func cmdWorktreeBackAt(origKey string, locations []LocationEntry, idx int, wtLoc LocationEntry) error {
	// Read the current main repo from the worktree's .git file on disk.
	mainRepo, err := readWorktreeMainRepo(wtLoc.Path)
	if err != nil {
		mainRepo = wtLoc.Git.MainRepo
		fmt.Fprintf(os.Stderr, "warning: could not read worktree main repo from disk, using history: %s\n", displayPath(mainRepo))
	}

	if _, err := os.Stat(mainRepo); err != nil {
		return fmt.Errorf("main repo %s no longer exists (worktree references stale path; the main repo may have been moved or deleted)", displayPath(mainRepo))
	}

	branch := wtLoc.Git.Branch

	if err := wt.IsClean(wtLoc.Path); err != nil {
		return err
	}

	// Determine branch relationship with main HEAD
	result, err := git.CompareBranches(mainRepo, branch, "HEAD")
	if err != nil {
		return err
	}

	switch result.Relation {
	case git.BranchRelationAIsAncestorOfB, git.BranchRelationSame:
		// CASE A: branch is ancestor of HEAD → already merged (existing behavior)
	case git.BranchRelationBIsAncestorOfA:
		// CASE B: HEAD is ancestor of branch → fast-forward possible
		confirmed, err := promptConfirm(fmt.Sprintf("branch %s is ahead, merge and remove?", branch))
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
		cmd := exec.Command("git", "-C", mainRepo, "merge", "--ff-only", branch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git merge --ff-only %s: %w\n%s", branch, err, out)
		}
	case git.BranchRelationDiverged:
		// CASE C: diverged - prompt and try rebase on worktree
		confirmed, err := promptConfirm(fmt.Sprintf("branch %s has diverged, rebase and merge?", branch))
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
		mainCommit, err := revParseCommit(mainRepo, "HEAD")
		if err != nil {
			return err
		}
		rebaseCmd := exec.Command("git", "-C", wtLoc.Path, "rebase", mainCommit)
		rebaseOut, rebaseErr := rebaseCmd.CombinedOutput()
		if rebaseErr != nil {
			abortCmd := exec.Command("git", "-C", wtLoc.Path, "rebase", "--abort")
			abortCmd.Run()
			return fmt.Errorf("rebase conflict: %w\n%s", rebaseErr, rebaseOut)
		}
		cmd := exec.Command("git", "-C", mainRepo, "merge", "--ff-only", branch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git merge --ff-only %s after rebase: %w\n%s", branch, err, out)
		}
	}

	if dryRun {
		fmt.Printf("dry-run: would remove worktree %s\n", displayPath(wtLoc.Path))
		return nil
	}

	cmd := exec.Command("git", "-C", mainRepo, "worktree", "remove", wtLoc.Path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove: %w\n%s", err, out)
	}

	cmd = exec.Command("git", "-C", mainRepo, "branch", "-D", branch)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch -D %s: %w\n%s", branch, err, out)
	}

	fmt.Printf("worktree removed: %s [branch: %s deleted]\n", displayPath(wtLoc.Path), branch)

	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}

	// Splice out the worktree entry at idx, preserving subsequent entries.
	newLocs := make([]LocationEntry, 0, len(locations)-1)
	newLocs = append(newLocs, locations[:idx]...)
	newLocs = append(newLocs, locations[idx+1:]...)
	hist[origKey] = newLocs
	if len(newLocs) <= 1 {
		delete(hist, origKey)
	}

	return saveHistory(hist, aliases)
}

func generateBranchName(basename, repoPath string) (string, error) {
	if !branchExists(basename, repoPath) {
		return basename, nil
	}
	date := time.Now().Format("2006-01-02")
	candidate := fmt.Sprintf("%s-%s", basename, date)
	if !branchExists(candidate, repoPath) {
		return candidate, nil
	}
	for i := 1; i < 100; i++ {
		candidate = fmt.Sprintf("%s-%s-%d", basename, date, i)
		if !branchExists(candidate, repoPath) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate unique branch name for %s after 99 attempts", basename)
}

func branchExists(branch, repoPath string) bool {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "refs/heads/"+branch)
	return cmd.Run() == nil
}

func revParseCommit(dir, ref string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", ref)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --verify %s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func stdinIsPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

func promptConfirm(msg string) (bool, error) {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	if !isTTY {
		if stdinIsPipe() {
			if !confirmFromStdin {
				return false, fmt.Errorf("stdin is not a terminal; pass --confirm-from-stdin to read confirmation from piped stdin")
			}
		} else {
			return false, fmt.Errorf("stdin is not a terminal; cannot prompt for confirmation")
		}
	}
	if isTTY {
		fmt.Fprintf(os.Stderr, "%s [Y/n]: ", msg)
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimSpace(strings.ToLower(line))
	switch line {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid input: %q (expected y/n)", strings.TrimSpace(line))
	}
}
