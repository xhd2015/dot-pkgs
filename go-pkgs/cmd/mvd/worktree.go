package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type worktreeInfo struct {
	path string
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

func isGitWorktree(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.Mode().IsRegular()
}

func listWorktrees(repoPath string) ([]worktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	all := parseWorktreeList(string(out))
	var wts []worktreeInfo
	for _, wt := range all {
		if isGitWorktree(wt.path) {
			wts = append(wts, wt)
		}
	}
	return wts, nil
}

func parseWorktreeList(output string) []worktreeInfo {
	var worktrees []worktreeInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	var current worktreeInfo
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			if current.path != "" {
				worktrees = append(worktrees, current)
			}
			current = worktreeInfo{path: line[len("worktree "):]}
		}
	}
	if current.path != "" {
		worktrees = append(worktrees, current)
	}
	return worktrees
}

func readWorktreeGitInfo(worktreePath string) (*GitInfo, error) {
	mainRepo, err := readWorktreeMainRepo(worktreePath)
	if err != nil {
		return nil, err
	}
	branch, err := readWorktreeBranch(worktreePath)
	if err != nil {
		return nil, err
	}
	return &GitInfo{
		Type:     "worktree",
		MainRepo: mainRepo,
		Branch:   branch,
	}, nil
}

func readWorktreeBranch(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func readWorktreeMainRepo(worktreePath string) (string, error) {
	gitFile := filepath.Join(worktreePath, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		return "", fmt.Errorf("read .git file: %w", err)
	}
	s := strings.TrimSpace(string(content))
	const prefix = "gitdir: "
	if !strings.HasPrefix(s, prefix) {
		return "", fmt.Errorf("unexpected .git file format in worktree %s", worktreePath)
	}
	gitdir := strings.TrimSpace(s[len(prefix):])
	// gitdir is <mainRepo>/.git/worktrees/<name>
	// go up: worktrees/<name> -> worktrees -> .git -> mainRepo
	mainRepo := filepath.Dir(filepath.Dir(filepath.Dir(gitdir)))
	return mainRepo, nil
}

func updateWorktreeGitFile(worktreePath, newRepo string) error {
	gitFile := filepath.Join(worktreePath, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		return fmt.Errorf("read .git: %w", err)
	}
	s := strings.TrimSpace(string(content))
	const prefix = "gitdir: "
	if !strings.HasPrefix(s, prefix) {
		return fmt.Errorf("unexpected .git file format in worktree %s", worktreePath)
	}
	gitdir := strings.TrimSpace(s[len(prefix):])
	// gitdir = <mainRepo>/.git/worktrees/<name>
	// Extract the worktree name and rebuild with new repo path
	name := filepath.Base(gitdir)
	newGitdir := filepath.Join(newRepo, ".git", "worktrees", name)
	newContent := prefix + newGitdir + "\n"
	return os.WriteFile(gitFile, []byte(newContent), 0644)
}
