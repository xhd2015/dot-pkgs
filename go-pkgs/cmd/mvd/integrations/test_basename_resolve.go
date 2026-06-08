package main

import (
	"os"
	"path/filepath"
)

func testWorktreeMoveByBasename(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	projectRoot := filepath.Join(work, "projects")
	mainRepo := filepath.Join(projectRoot, "myrepo")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", mainRepo)
	assertHistoryChain(t, mainRepo, mainRepo)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	wtDir := filepath.Join(work, "feature")
	out := runMvdOk(t, "-w", "myrepo", wtDir)

	assertContains(t, out, "worktree created:")
	assertHistoryChain(t, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, mainRepo, 1, mainRepo, "feature")
	assertFileExists(t, filepath.Join(wtDir, ".git"))
}

func testClearByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", dir)
	assertHistoryChain(t, dir, dir)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	out := runMvdOk(t, "--clear", "myproject")
	assertContains(t, out, "cleared history")
	assertHistoryNil(t)
}

func testRebaseByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	newBase := filepath.Join(work, "newbase")
	assertNoErr(t, os.MkdirAll(dir, 0755))
	assertNoErr(t, os.MkdirAll(newBase, 0755))

	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", dir)
	assertHistoryChain(t, dir, dir)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	out := runMvdOk(t, "--rebase", "myproject", newBase)
	assertContains(t, out, "rebased:")
	assertHistoryChain(t, newBase, newBase, dir)
}

func testListByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", dir)
	assertHistoryChain(t, dir, dir)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	out := runMvdOk(t, "--list", "myproject")
	assertContains(t, out, "myproject")
	assertContains(t, out, "(original)")
}
