package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func testMoveWorktreeWithoutWFlagShouldDoSimpleMove(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature-wt")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	wtDst := filepath.Join(work, "feature-wt-moved")
	out := runMvdOk(t, wtDir, wtDst)

	assertNotContains(t, out, "worktree created:")
	assertNotContains(t, out, "worktree add")
	assertFileExists(t, wtDst)
	assertFileNotExists(t, wtDir)

	assertFileExists(t, filepath.Join(wtDst, ".git"))
	gitContent, err := os.ReadFile(filepath.Join(wtDst, ".git"))
	assertNoErr(t, err)
	assertContains(t, string(gitContent), mainRepo)
}

func testMoveNestedWorktreeWithoutWFlag(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wt1Dir := filepath.Join(work, "readonly-master")
	runGit(t, mainRepo, "worktree", "add", wt1Dir)

	wt2Dir := filepath.Join(work, "pricing")
	cmd := exec.Command("git", "-C", wt1Dir, "worktree", "add", wt2Dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git worktree add from worktree1: %v\n%s", err, out)
	}

	assertFileExists(t, filepath.Join(wt2Dir, ".git"))

	wt2Dst := filepath.Join(work, "pricing-moved")
	mvdOut := runMvdOk(t, wt2Dir, wt2Dst)

	assertNotContains(t, mvdOut, "worktree created:")
	assertFileExists(t, wt2Dst)
	assertFileNotExists(t, wt2Dir)

	assertFileExists(t, filepath.Join(wt2Dst, ".git"))
	gitContent, err := os.ReadFile(filepath.Join(wt2Dst, ".git"))
	assertNoErr(t, err)
	assertContains(t, string(gitContent), mainRepo)
}
