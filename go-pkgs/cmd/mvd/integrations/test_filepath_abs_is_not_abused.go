package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func setupLlsConfig(t *T, projectRoot string) (envOverrides []string) {
	home := filepath.Join(tmpBase, t.name, ".lls-home")
	configDir := filepath.Join(home, "Library", "Application Support", "lls")
	t.Helper()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir lls config dir: %v", err)
	}
	writeFile(t, filepath.Join(configDir, "config.json"), `{"envs":["X"]}`)
	return []string{
		"HOME=" + home,
		"X=" + projectRoot,
	}
}

func runMvdWithLlsConfig(t *T, envOverrides []string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(mvdBin, args...)
	baseEnv := append(os.Environ(), "MVD_DEBUG_CONFIG_HOME="+configHome(t))
	cmd.Env = append(baseEnv, envOverrides...)
	out, err := cmd.CombinedOutput()
	t.Logf("%s", string(out))
	if err != nil {
		return string(out), fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func testClearWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", dir)
	assertHistoryChain(t, dir, dir)

	out, err := runMvdWithLlsConfig(t, env, "--clear", "$X/myproject")
	assertNoErr(t, err)
	assertContains(t, out, "cleared history")
	assertHistoryNil(t)
}

func testBackWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	dst := filepath.Join(work, "moved")
	assertNoErr(t, os.MkdirAll(dir, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", dir)
	movedPath := filepath.Join(dst, "myproject")
	runMvdWithLlsConfig(t, env, dir, dst)
	assertHistoryChain(t, dir, dir, movedPath)

	out, err := runMvdWithLlsConfig(t, env, "--back", "$X/myproject")
	assertNoErr(t, err)
	assertContains(t, out, "moved back")
	assertFileExists(t, dir)
}

func testListWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", dir)

	out, err := runMvdWithLlsConfig(t, env, "--list", "$X/myproject")
	assertNoErr(t, err)
	assertContains(t, out, "myproject")
}

func testRebaseWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	newBase := filepath.Join(projectRoot, "rebased-project")
	assertNoErr(t, os.MkdirAll(dir, 0755))
	assertNoErr(t, os.MkdirAll(newBase, 0755))

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", dir)

	out, err := runMvdWithLlsConfig(t, env, "--rebase", "$X/myproject", newBase)
	assertNoErr(t, err)
	assertContains(t, out, "rebased:")
	assertHistoryChain(t, newBase, newBase, dir)
}

func testWorktreeMoveWithDollarExpansion(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	mainRepo := filepath.Join(projectRoot, "myrepo")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", mainRepo)

	wtDir := filepath.Join(work, "feature")
	out, err := runMvdWithLlsConfig(t, env, "-w", "$X/myrepo", wtDir)
	assertNoErr(t, err)
	assertContains(t, out, "worktree created:")
	assertHistoryChain(t, mainRepo, mainRepo, wtDir)
	assertFileExists(t, filepath.Join(wtDir, ".git"))
}

func testWhichWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", dir)

	out, err := runMvdWithLlsConfig(t, env, "--which", "$X/myproject")
	assertNoErr(t, err)
	assertContains(t, out, "(local)")
}

func testAddWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	env := setupLlsConfig(t, projectRoot)

	out, err := runMvdWithLlsConfig(t, env, "--add", "$X/myproject")
	assertNoErr(t, err)
	assertContains(t, out, "added:")
	assertHistoryChain(t, dir, dir)
}

func testAddNonExistentFails(t *T) {
	work := testDir(t, "work")
	nonExistent := filepath.Join(work, "no-such-dir")

	out := runMvdErr(t, "--add", nonExistent)
	assertContains(t, out, "does not exist")
	assertHistoryNil(t)
}

func testMoveDefaultWithDollarExpansion(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "myproject")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	env := setupLlsConfig(t, projectRoot)

	runMvdWithLlsConfig(t, env, "--add", src)

	out, err := runMvdWithLlsConfig(t, env, "$X/myproject", dst)
	assertNoErr(t, err)
	assertContains(t, out, "moved:")
	finalPath := filepath.Join(dst, "myproject")
	assertHistoryChain(t, src, src, finalPath)
	assertFileExists(t, filepath.Join(finalPath, "f.txt"))
}
