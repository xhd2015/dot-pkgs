package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func initGitRepo(t *testing.T, path string) {
	t.Helper()
	runGit(t, path, "init")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	// Create a commit so worktree add works
	if err := os.WriteFile(filepath.Join(path, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", "init")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func TestIsGitRepoTrue(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initGitRepo(t, dir)
	if !isGitRepo(dir) {
		t.Fatal("expected isGitRepo to return true for git repo")
	}
}

func TestIsGitRepoFalseForNonGitDir(t *testing.T) {
	dir := t.TempDir()
	if isGitRepo(dir) {
		t.Fatal("expected isGitRepo to return false for non-git dir")
	}
}

func TestIsGitWorktreeTrue(t *testing.T) {
	skipIfNoGit(t)
	mainRepo := t.TempDir()
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(t.TempDir(), "feature")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	if !isGitWorktree(wtDir) {
		t.Fatal("expected isGitWorktree to return true for worktree")
	}
}

func TestIsGitWorktreeFalseForRegularRepo(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initGitRepo(t, dir)
	if isGitWorktree(dir) {
		t.Fatal("expected isGitWorktree to return false for regular git repo")
	}
}

func TestIsGitWorktreeFalseForNonGitDir(t *testing.T) {
	dir := t.TempDir()
	if isGitWorktree(dir) {
		t.Fatal("expected isGitWorktree to return false for non-git dir")
	}
}

func TestListWorktreesExcludesMainRepo(t *testing.T) {
	skipIfNoGit(t)
	mainRepo := t.TempDir()
	initGitRepo(t, mainRepo)

	wtDir1 := filepath.Join(t.TempDir(), "feature1")
	wtDir2 := filepath.Join(t.TempDir(), "feature2")
	runGit(t, mainRepo, "worktree", "add", wtDir1)
	runGit(t, mainRepo, "worktree", "add", wtDir2)

	wts, err := listWorktrees(mainRepo)
	if err != nil {
		t.Fatalf("listWorktrees: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d: %+v", len(wts), wts)
	}
	mainReal, _ := filepath.EvalSymlinks(mainRepo)
	for _, wt := range wts {
		if wt.path == mainReal {
			t.Fatal("main repo should not be listed as worktree")
		}
	}
}

func TestListWorktreesNoWorktrees(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initGitRepo(t, dir)

	wts, err := listWorktrees(dir)
	if err != nil {
		t.Fatalf("listWorktrees: %v", err)
	}
	if len(wts) != 0 {
		t.Fatalf("expected 0 worktrees, got %d", len(wts))
	}
}

func resolvePath(p string) string {
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p
	}
	return r
}

func TestReadWorktreeMainRepo(t *testing.T) {
	skipIfNoGit(t)
	mainRepo := t.TempDir()
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(t.TempDir(), "feature")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	got, err := readWorktreeMainRepo(wtDir)
	if err != nil {
		t.Fatalf("readWorktreeMainRepo: %v", err)
	}
	if resolvePath(got) != resolvePath(mainRepo) {
		t.Fatalf("expected main repo %s, got %s", resolvePath(mainRepo), resolvePath(got))
	}
}

func TestUpdateWorktreeGitFile(t *testing.T) {
	skipIfNoGit(t)
	mainRepo := t.TempDir()
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(t.TempDir(), "feature")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	newMainRepo := filepath.Join(t.TempDir(), "moved-repo")
	if err := os.Rename(mainRepo, newMainRepo); err != nil {
		t.Fatalf("rename: %v", err)
	}

	if err := updateWorktreeGitFile(wtDir, newMainRepo); err != nil {
		t.Fatalf("updateWorktreeGitFile: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(wtDir, ".git"))
	if err != nil {
		t.Fatalf("read .git: %v", err)
	}
	if !strings.Contains(string(content), newMainRepo) {
		t.Fatalf("expected .git file to contain %s, got %s", newMainRepo, content)
	}
	if strings.Contains(string(content), mainRepo) {
		t.Fatalf(".git file should not contain old repo path %s: %s", mainRepo, content)
	}
}

func TestCmdMoveWorktree(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature-wt")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	// Now move the worktree
	wtDst := filepath.Join(work, "feature-wt-moved")
	output := captureStdout(t, func() {
		if err := cmdMove(wtDir, wtDst); err != nil {
			t.Fatalf("cmdMove: %v", err)
		}
	})

	if !strings.Contains(output, "moved worktree:") {
		t.Fatalf("expected 'moved worktree:' output, got: %s", output)
	}
	if !pathExists(wtDst) {
		t.Fatal("worktree was not moved to destination")
	}
	if pathExists(wtDir) {
		t.Fatal("old worktree path still exists")
	}

	// Verify .git file in moved worktree still points correctly
	mainRepoFromWt, err := readWorktreeMainRepo(wtDst)
	if err != nil {
		t.Fatalf("readWorktreeMainRepo: %v", err)
	}
	if resolvePath(mainRepoFromWt) != resolvePath(mainRepo) {
		t.Fatalf("worktree .git should reference main repo %s, got %s", resolvePath(mainRepo), resolvePath(mainRepoFromWt))
	}
}

func TestCmdMoveRepoWithWorktrees(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir1 := filepath.Join(work, "feature1")
	wtDir2 := filepath.Join(work, "feature2")
	runGit(t, mainRepo, "worktree", "add", wtDir1)
	runGit(t, mainRepo, "worktree", "add", wtDir2)

	// Move the main repo
	mainDst := filepath.Join(work, "main-moved")
	output := captureStdout(t, func() {
		if err := cmdMove(mainRepo, mainDst); err != nil {
			t.Fatalf("cmdMove: %v", err)
		}
	})

	if !strings.Contains(output, "moved:") {
		t.Fatalf("expected 'moved:' output, got: %s", output)
	}
	if !strings.Contains(output, "updated worktree:") {
		t.Fatalf("expected 'updated worktree:' output, got: %s", output)
	}
	if !pathExists(mainDst) {
		t.Fatal("main repo was not moved")
	}
	if pathExists(mainRepo) {
		t.Fatal("old main repo path still exists")
	}

	// Verify worktree .git files point to new location
	for _, wtDir := range []string{wtDir1, wtDir2} {
		repo, err := readWorktreeMainRepo(wtDir)
		if err != nil {
			t.Fatalf("readWorktreeMainRepo(%s): %v", wtDir, err)
		}
		if resolvePath(repo) != resolvePath(mainDst) {
			t.Fatalf("worktree %s: expected main repo %s, got %s", wtDir, resolvePath(mainDst), resolvePath(repo))
		}
	}
}

func TestCmdBackWorktree(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtOriginal := filepath.Join(work, "feature-wt")
	runGit(t, mainRepo, "worktree", "add", wtOriginal)

	// First move forward
	wtIntermediate := filepath.Join(work, "feature-moved")
	if err := cmdMove(wtOriginal, wtIntermediate); err != nil {
		t.Fatalf("cmdMove forward: %v", err)
	}

	// Then move back
	output := captureStdout(t, func() {
		if err := cmdBack(wtIntermediate); err != nil {
			t.Fatalf("cmdBack: %v", err)
		}
	})

	if !strings.Contains(output, "moved worktree back:") {
		t.Fatalf("expected 'moved worktree back:' output, got: %s", output)
	}
	if !pathExists(wtOriginal) {
		t.Fatal("worktree was not moved back")
	}
}

func TestCmdBackRepoWithWorktrees(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	mainRepo := filepath.Join(work, "main")
	mustMkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature1")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	// Move main repo forward
	mainDst := filepath.Join(work, "main-moved")
	if err := cmdMove(mainRepo, mainDst); err != nil {
		t.Fatalf("cmdMove forward: %v", err)
	}

	// Move it back
	output := captureStdout(t, func() {
		if err := cmdBack(mainDst); err != nil {
			t.Fatalf("cmdBack: %v", err)
		}
	})

	if !strings.Contains(output, "moved back:") {
		t.Fatalf("expected 'moved back:' output, got: %s", output)
	}
	if !strings.Contains(output, "updated worktree:") {
		t.Fatalf("expected 'updated worktree:' output, got: %s", output)
	}
	if !pathExists(mainRepo) {
		t.Fatal("main repo was not moved back")
	}

	// Verify worktree .git file points to original location again
	repo, err := readWorktreeMainRepo(wtDir)
	if err != nil {
		t.Fatalf("readWorktreeMainRepo: %v", err)
	}
	if resolvePath(repo) != resolvePath(mainRepo) {
		t.Fatalf("worktree: expected main repo %s, got %s", resolvePath(mainRepo), resolvePath(repo))
	}
}

func TestMoveDirPlainDir(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	srcDir := filepath.Join(work, "plain-dir")
	dstDir := filepath.Join(work, "plain-dir-moved")
	mustMkdirAll(t, srcDir)

	if err := moveDir(srcDir, dstDir); err != nil {
		t.Fatalf("moveDir: %v", err)
	}

	if !pathExists(dstDir) {
		t.Fatal("dir was not moved")
	}
	if pathExists(srcDir) {
		t.Fatal("old dir still exists")
	}
}

func TestMoveDirRepoWithoutWorktrees(t *testing.T) {
	skipIfNoGit(t)
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(work, "repo")
	mustMkdirAll(t, repo)
	initGitRepo(t, repo)

	repoDst := filepath.Join(work, "repo-moved")

	output := captureStdout(t, func() {
		if err := moveDir(repo, repoDst); err != nil {
			t.Fatalf("moveDir: %v", err)
		}
	})

	if !pathExists(repoDst) {
		t.Fatal("repo was not moved")
	}
	if pathExists(repo) {
		t.Fatal("old repo still exists")
	}
	// No worktree updates expected
	if strings.Contains(output, "updated worktree:") {
		t.Fatalf("expected no worktree updates, got: %s", output)
	}
}

func TestParseWorktreeList(t *testing.T) {
	input := `worktree /path/to/repo
HEAD 1234567890abcdef
branch refs/heads/main

worktree /path/to/repo/feature
HEAD 1234567890abcdef
branch refs/heads/feature
`

	wts := parseWorktreeList(input)
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}
	if wts[0].path != "/path/to/repo" {
		t.Fatalf("expected first worktree to be /path/to/repo, got %s", wts[0].path)
	}
	if wts[1].path != "/path/to/repo/feature" {
		t.Fatalf("expected second worktree to be /path/to/repo/feature, got %s", wts[1].path)
	}
}

func TestParseWorktreeListEmpty(t *testing.T) {
	wts := parseWorktreeList("")
	if len(wts) != 0 {
		t.Fatalf("expected 0 worktrees, got %d", len(wts))
	}
}

func TestParseWorktreeListWithBare(t *testing.T) {
	input := `worktree /path/to/bare
bare

worktree /path/to/repo
HEAD abc
branch refs/heads/main
`
	wts := parseWorktreeList(input)
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}
}

func TestCmdMoveNewDir(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	srcDir := filepath.Join(work, "new-dir")
	dstDir := filepath.Join(work, "new-dir-moved")
	mustMkdirAll(t, srcDir)

	output := captureStdout(t, func() {
		if err := cmdMove(srcDir, dstDir); err != nil {
			t.Fatalf("cmdMove: %v", err)
		}
	})

	if !bytes.Contains([]byte(output), []byte("moved:")) {
		t.Fatalf("expected 'moved:' in output, got: %s", output)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(hist) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(hist))
	}
	locs := hist[srcDir]
	if len(locs) != 2 || locs[0] != srcDir || locs[1] != dstDir {
		t.Fatalf("expected history [%s %s], got %v", srcDir, dstDir, locs)
	}
}
