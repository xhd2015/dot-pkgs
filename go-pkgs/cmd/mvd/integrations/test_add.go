package main

import (
	"os"
	"path/filepath"
)

func testAdd(t *T) {
	work := testDir(t, "work")
	dir := filepath.Join(work, "tracked")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	assertHistoryNil(t)
	runMvdOk(t, "--add", dir)
	assertHistoryChain(t, dir, dir)
}

func testAddDuplicate(t *T) {
	work := testDir(t, "work")
	dir := filepath.Join(work, "tracked")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	runMvdOk(t, "--add", dir)
	out := runMvdOk(t, "--add", dir)
	assertContains(t, out, "already recorded")
	assertHistoryLen(t, 1)
}
