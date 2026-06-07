package main

import (
	"os"
	"path/filepath"
)

func testRemove(t *T) {
	work := testDir(t, "work")
	dir := filepath.Join(work, "tracked")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	runMvdOk(t, "--add", dir)
	runMvdOk(t, "--rm", dir)
	assertHistoryNil(t)
}

func testRemoveForce(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))

	runMvdOk(t, src, d1)

	out := runMvdOk(t, "--rm", "-f", src)
	assertContains(t, out, "will clear")
	assertHistoryNil(t)
}

func testRemoveNoForceWithHistory(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))

	runMvdOk(t, src, d1)

	out := runMvdErr(t, "--rm", src)
	assertContains(t, out, "has movement history")
	assertHistoryContainsKey(t, src)
}
