package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

	if !isGitRepo(srcAbs) {
		return fmt.Errorf("%s is not a git repository", displayPath(srcAbs))
	}

	branch, err := generateBranchName(filepath.Base(dstAbs), srcAbs)
	if err != nil {
		return err
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
	mainRepo := last.Git.MainRepo
	branch := last.Git.Branch

	if err := checkWorktreeClean(last.Path); err != nil {
		return err
	}

	if err := checkBranchMerged(branch, mainRepo); err != nil {
		return err
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

func checkWorktreeClean(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return fmt.Errorf(
			"worktree %s has uncommitted changes:\n"+
				"  commit or stash them before moving back\n"+
				"  %s",
			displayPath(worktreePath),
			string(out),
		)
	}
	return nil
}

func checkBranchMerged(branch, repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "merge-base", "--is-ancestor", branch, "HEAD")
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf(
				"branch %s is not merged into %s HEAD:\n"+
					"  merge the branch first to avoid data loss, then run --back again",
				branch, displayPath(repoPath),
			)
		}
		return fmt.Errorf("git merge-base: %w", err)
	}
	return nil
}
