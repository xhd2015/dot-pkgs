package main

import (
	"os"
	"path/filepath"
)

func testBack(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "src")

	runMvdOk(t, "--back", p)
	assertHistoryChain(t, src, src)
	assertFileExists(t, filepath.Join(src, "f.txt"))
	assertFileNotExists(t, p)
}

func testBackAtOrigin(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "src")

	runMvdOk(t, "--back", p)
	out := runMvdOk(t, "--back", src)
	assertContains(t, out, "nothing to move back")
	assertHistoryChain(t, src, src)
}

func testBackByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst := filepath.Join(work, "scratch")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "kool")

	out := runMvdOk(t, "--back", "kool")
	assertContains(t, out, "moved back")
	assertHistoryChain(t, src, src)
	assertFileExists(t, src)
	assertFileNotExists(t, p)
}

func testMultiStepBack(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	d2 := filepath.Join(work, "d2")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))
	assertNoErr(t, os.MkdirAll(d2, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")

	runMvdOk(t, p1, d2)
	p2 := filepath.Join(d2, "src")

	assertHistoryChain(t, src, src, p1, p2)

	runMvdOk(t, "--back", p2)
	assertHistoryChain(t, src, src, p1)
	assertFileExists(t, p1)

	runMvdOk(t, "--back", p1)
	assertHistoryChain(t, src, src)
	assertFileExists(t, src)

	out := runMvdOk(t, "--back", src)
	assertContains(t, out, "nothing to move back")
}
