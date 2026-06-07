package main

import (
	"os"
	"path/filepath"
)

func testClear(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	out := runMvdOk(t, "--clear", src)
	assertContains(t, out, "cleared history")
	assertHistoryNil(t)
}
