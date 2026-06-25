package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wt "github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

type worktreeInfo struct {
	path string
}

func isGitRepo(path string) bool {
	return wt.IsMainRepo(path)
}

func isGitWorktree(path string) bool {
	return wt.IsLinked(path)
}

func listWorktrees(repoPath string) ([]worktreeInfo, error) {
	entries, err := wt.ListLinked(repoPath)
	if err != nil {
		return nil, err
	}
	wts := make([]worktreeInfo, len(entries))
	for i, e := range entries {
		wts[i] = worktreeInfo{path: e.Path}
	}
	return wts, nil
}

func readWorktreeGitInfo(worktreePath string) (*GitInfo, error) {
	mainRepo, err := wt.ReadMainRepo(worktreePath)
	if err != nil {
		return nil, err
	}
	branch, err := wt.ReadBranch(worktreePath)
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
	return wt.ReadBranch(worktreePath)
}

func readWorktreeMainRepo(worktreePath string) (string, error) {
	return wt.ReadMainRepo(worktreePath)
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
	name := filepath.Base(gitdir)
	newGitdir := filepath.Join(newRepo, ".git", "worktrees", name)
	newContent := prefix + newGitdir + "\n"
	return os.WriteFile(gitFile, []byte(newContent), 0644)
}