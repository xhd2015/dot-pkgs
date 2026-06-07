package main

import (
	"os"
	"path/filepath"
)

func testListAll(t *T) {
	work := testDir(t, "work")
	s1 := filepath.Join(work, "proj1")
	s2 := filepath.Join(work, "proj2")
	assertNoErr(t, os.MkdirAll(s1, 0755))
	assertNoErr(t, os.MkdirAll(s2, 0755))

	runMvdOk(t, "--add", s1)
	runMvdOk(t, "--add", s2)

	out := runMvdOk(t, "--list")
	assertContains(t, out, s1)
	assertContains(t, out, s2)
}

func testListSingle(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "src")
	assertFileExists(t, p)

	out := runMvdOk(t, "--list", src)
	assertContains(t, out, "(original)")
	assertContains(t, out, "*")
}
