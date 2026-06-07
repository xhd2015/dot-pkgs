package main

import (
	"path/filepath"
)

func testNonExistentSrc(t *T) {
	work := testDir(t, "work")
	nonexistent := filepath.Join(work, "no-such-dir")
	dst := filepath.Join(work, "dst")

	out := runMvdErr(t, nonexistent, dst)
	assertContains(t, out, "no such file or directory")
}
