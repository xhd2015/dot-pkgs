package main

import (
	"os"
	"path/filepath"
)

func testMoveWorktreeWithWFlagShouldRunGitWorktreeAdd(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature-wt")
	out := runMvdOk(t, "-w", mainRepo, wtDir)

	assertContains(t, out, "worktree created:")
	assertContains(t, out, "[branch: feature-wt]")
	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertFileExists(t, filepath.Join(wtDir, "README.md"))

	assertHistoryChain(t, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, mainRepo, 1, mainRepo, "feature-wt")
}
