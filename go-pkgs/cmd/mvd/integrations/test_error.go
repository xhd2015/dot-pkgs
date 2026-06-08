package main

import (
	"os"
	"path/filepath"
)

func testNonExistentSrc(t *T) {
	work := testDir(t, "work")
	nonexistent := filepath.Join(work, "no-such-dir")
	dst := filepath.Join(work, "dst")

	out := runMvdErr(t, nonexistent, dst)
	assertContains(t, out, "does not exist")
}

func testMoveNonExistentBasename(t *T) {
	work := testDir(t, "work")
	dst := filepath.Join(work, "dst")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(dst, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	out := runMvdErr(t, "git-ops", dst)
	assertContains(t, out, "git-ops does not exist, no configured project match basename or alias git-ops")
}
