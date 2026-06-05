package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorktreeMoveCreatesWorktree(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	output := captureStdout(t, func() {
		if err := cmdWorktreeMove(mainRepo, wtDir); err != nil {
			t.Fatalf("cmdWorktreeMove: %v", err)
		}
	})

	if !strings.Contains(output, "worktree created:") {
		t.Fatalf("expected 'worktree created:' output, got: %s", output)
	}
	if !strings.Contains(output, "[branch: feature]") {
		t.Fatalf("expected branch name in output, got: %s", output)
	}
	if !pathExists(wtDir) {
		t.Fatal("worktree directory was not created")
	}
	if !isGitWorktree(wtDir) {
		t.Fatal("destination is not a git worktree")
	}

	mainRepoFromWt, err := readWorktreeMainRepo(wtDir)
	if err != nil {
		t.Fatalf("readWorktreeMainRepo: %v", err)
	}
	if resolvePath(mainRepoFromWt) != resolvePath(mainRepo) {
		t.Fatalf("worktree .git should reference main repo %s, got %s", resolvePath(mainRepo), resolvePath(mainRepoFromWt))
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[mainRepo]
	if len(locs) != 2 {
		t.Fatalf("expected 2 locations, got %d: %+v", len(locs), locs)
	}
	if locs[0].Path != mainRepo {
		t.Fatalf("expected first location %s, got %s", mainRepo, locs[0].Path)
	}
	if locs[1].Path != wtDir {
		t.Fatalf("expected second location %s, got %s", wtDir, locs[1].Path)
	}
	gitInfo := locs[1].Git
	if gitInfo == nil || gitInfo.Type != "worktree" {
		t.Fatalf("expected git.type=worktree, got %+v", gitInfo)
	}
	if gitInfo.MainRepo != mainRepo {
		t.Fatalf("expected main_repo %s, got %s", mainRepo, gitInfo.MainRepo)
	}
	if gitInfo.Branch != "feature" {
		t.Fatalf("expected branch 'feature', got %s", gitInfo.Branch)
	}
}

func TestWorktreeMoveNonGitSrcFails(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	nonGitDir := filepath.Join(work, "not-a-repo")
	mustMkdirAll(t, nonGitDir)

	wtDir := filepath.Join(work, "feature")
	err := cmdWorktreeMove(nonGitDir, wtDir)
	if err == nil {
		t.Fatal("expected error for non-git src")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateBranchNameUnique(t *testing.T) {
	skipIfNoGit(t)
	repo := t.TempDir()
	initGitRepo(t, repo)

	branch, err := generateBranchName("new-feature", repo)
	if err != nil {
		t.Fatalf("generateBranchName: %v", err)
	}
	if branch != "new-feature" {
		t.Fatalf("expected 'new-feature', got %s", branch)
	}
}

func TestGenerateBranchNameCollision(t *testing.T) {
	skipIfNoGit(t)
	repo := t.TempDir()
	initGitRepo(t, repo)

	runGit(t, repo, "branch", "existing")

	branch, err := generateBranchName("existing", repo)
	if err != nil {
		t.Fatalf("generateBranchName: %v", err)
	}
	if branch == "existing" {
		t.Fatal("expected a different branch name when 'existing' already exists")
	}
	if !strings.HasPrefix(branch, "existing-") {
		t.Fatalf("expected branch to start with 'existing-', got %s", branch)
	}
}

func TestCheckWorktreeCleanClean(t *testing.T) {
	skipIfNoGit(t)
	repo := t.TempDir()
	initGitRepo(t, repo)

	wtDir := filepath.Join(t.TempDir(), "clean-wt")
	runGit(t, repo, "worktree", "add", wtDir)

	if err := checkWorktreeClean(wtDir); err != nil {
		t.Fatalf("expected clean worktree, got error: %v", err)
	}
}

func TestCheckWorktreeCleanDirty(t *testing.T) {
	skipIfNoGit(t)
	repo := t.TempDir()
	initGitRepo(t, repo)

	wtDir := filepath.Join(t.TempDir(), "dirty-wt")
	runGit(t, repo, "worktree", "add", wtDir)

	if err := os.WriteFile(filepath.Join(wtDir, "dirty-file"), []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("write dirty-file: %v", err)
	}

	if err := checkWorktreeClean(wtDir); err == nil {
		t.Fatal("expected error for dirty worktree")
	} else if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckBranchMergedTrue(t *testing.T) {
	skipIfNoGit(t)
	repo := t.TempDir()
	initGitRepo(t, repo)

	runGit(t, repo, "checkout", "-b", "feature")
	runGit(t, repo, "checkout", "-")
	runGit(t, repo, "merge", "feature")

	if err := checkBranchMerged("feature", repo); err != nil {
		t.Fatalf("expected merged branch, got error: %v", err)
	}
}

func TestCheckBranchMergedFalse(t *testing.T) {
	skipIfNoGit(t)
	repo := t.TempDir()
	initGitRepo(t, repo)

	runGit(t, repo, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(repo, "new-file"), []byte("content"), 0644); err != nil {
		t.Fatalf("write new-file: %v", err)
	}
	runGit(t, repo, "add", "new-file")
	runGit(t, repo, "commit", "-m", "feature work")
	runGit(t, repo, "checkout", "-")

	if err := checkBranchMerged("feature", repo); err == nil {
		t.Fatal("expected error for unmerged branch")
	} else if !strings.Contains(err.Error(), "not merged") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorktreeBackDirtyFails(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	if err := cmdWorktreeMove(mainRepo, wtDir); err != nil {
		t.Fatalf("cmdWorktreeMove: %v", err)
	}

	if err := os.WriteFile(filepath.Join(wtDir, "dirty-file"), []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("write dirty-file: %v", err)
	}

	err := cmdBack(wtDir)
	if err == nil {
		t.Fatal("expected error for dirty worktree on --back")
	}
	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pathExists(wtDir) {
		t.Fatal("worktree should not be deleted when dirty")
	}
}

func TestWorktreeBackUnmergedFails(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	if err := cmdWorktreeMove(mainRepo, wtDir); err != nil {
		t.Fatalf("cmdWorktreeMove: %v", err)
	}

	if err := os.WriteFile(filepath.Join(wtDir, "feature-work"), []byte("work"), 0644); err != nil {
		t.Fatalf("write feature-work: %v", err)
	}
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature commit")

	err := cmdBack(wtDir)
	if err == nil {
		t.Fatal("expected error for unmerged branch on --back")
	}
	if !strings.Contains(err.Error(), "not merged") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pathExists(wtDir) {
		t.Fatal("worktree should not be deleted when unmerged")
	}
}

func TestWorktreeBackCleanMergedSucceeds(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	if err := cmdWorktreeMove(mainRepo, wtDir); err != nil {
		t.Fatalf("cmdWorktreeMove: %v", err)
	}

	if err := os.WriteFile(filepath.Join(wtDir, "feature-work"), []byte("work"), 0644); err != nil {
		t.Fatalf("write feature-work: %v", err)
	}
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature commit")

	runGit(t, mainRepo, "merge", "feature")

	output := captureStdout(t, func() {
		if err := cmdBack(wtDir); err != nil {
			t.Fatalf("cmdBack: %v", err)
		}
	})

	if !strings.Contains(output, "worktree removed:") {
		t.Fatalf("expected 'worktree removed:' output, got: %s", output)
	}
	if !strings.Contains(output, "branch: feature deleted") {
		t.Fatalf("expected branch deletion message, got: %s", output)
	}
	if pathExists(wtDir) {
		t.Fatal("worktree should be deleted after clean+merged back")
	}

	cmd := exec.Command("git", "-C", mainRepo, "rev-parse", "--verify", "refs/heads/feature")
	if cmd.Run() == nil {
		t.Fatal("branch 'feature' should have been deleted")
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[mainRepo]; ok {
		t.Fatalf("expected history entry for %s to be removed after worktree back", mainRepo)
	}
}

func TestRunWorktreeFlagDispatches(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	output := captureStdout(t, func() {
		if err := run([]string{"-w", mainRepo, wtDir}); err != nil {
			t.Fatalf("run -w: %v", err)
		}
	})

	if !strings.Contains(output, "worktree created:") {
		t.Fatalf("expected 'worktree created:' output, got: %s", output)
	}
	if !pathExists(wtDir) {
		t.Fatal("worktree directory was not created")
	}
}
