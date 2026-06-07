package main

import (
	"os"
	"path/filepath"
	"strings"
)

func testWorktreeMove(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")

	out := runMvdOk(t, "-w", mainRepo, wtDir)
	assertContains(t, out, "worktree created:")
	assertContains(t, out, "[branch: feature]")
	assertFileExists(t, filepath.Join(wtDir, ".git"))

	assertHistoryChain(t, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, mainRepo, 1, mainRepo, "feature")
}

func testWorktreeNonGitSrc(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	nonGit := filepath.Join(work, "not-a-repo")
	assertNoErr(t, os.MkdirAll(nonGit, 0755))
	wtDir := filepath.Join(work, "feature")

	out := runMvdErr(t, "-w", nonGit, wtDir)
	assertContains(t, out, "not a git repository")
	assertHistoryNil(t)
}

func testWorktreeBackDirty(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	runMvdOk(t, "-w", mainRepo, wtDir)

	writeFile(t, filepath.Join(wtDir, "dirty-file"), "uncommitted")

	out := runMvdErr(t, "--back", wtDir)
	assertContains(t, out, "uncommitted changes")
	assertFileExists(t, wtDir)
}

func testWorktreeBackUnmerged(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	runMvdOk(t, "-w", mainRepo, wtDir)

	writeFile(t, filepath.Join(wtDir, "feature-work"), "work")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	out := runMvdErr(t, "--back", wtDir)
	assertContains(t, out, "not merged")
	assertFileExists(t, wtDir)
}

func testWorktreeBackSuccess(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	runMvdOk(t, "-w", mainRepo, wtDir)

	writeFile(t, filepath.Join(wtDir, "feature-work"), "work")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	runGit(t, mainRepo, "merge", "feature")

	out := runMvdOk(t, "--back", wtDir)
	assertContains(t, out, "worktree removed:")
	assertContains(t, out, "branch: feature deleted")
	assertFileNotExists(t, wtDir)
	assertHistoryNil(t)
}

func testWorktreeBranchCollision(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	runGit(t, mainRepo, "branch", "myfeature")

	wtDir := filepath.Join(work, "myfeature")
	out := runMvdOk(t, "-w", mainRepo, wtDir)

	assertContains(t, out, "worktree created:")
	assertNotContains(t, out, "[branch: myfeature]")

	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertHistoryChain(t, mainRepo, mainRepo, wtDir)

	h := readHistory(t)
	proj := h.Projects[mainRepo]
	branch := proj.Locations[1].Git.Branch
	if branch == "myfeature" {
		t.Fatalf("branch name should not be 'myfeature' (collision)")
	}
	if !strings.HasPrefix(branch, "myfeature-") {
		t.Fatalf("branch name should start with 'myfeature-', got %q", branch)
	}
}
