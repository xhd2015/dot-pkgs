package main

import (
	"os"
	"path/filepath"
)

func testRebase(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	newBase := filepath.Join(work, "rebased")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")

	runMvdOk(t, "--rebase", src, newBase)
	assertHistoryChain(t, newBase, newBase, src, p1)
}
