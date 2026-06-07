package main

import (
	"os"
	"path/filepath"
)

func testBasicMove(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	assertHistoryNil(t)
	runMvdOk(t, src, dst)
	assertHistoryChain(t, src, src, dst)
	assertFileExists(t, filepath.Join(dst, "f.txt"))
	assertFileNotExists(t, src)
}

func testMoveToExistingDir(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "mysrc")
	dst := filepath.Join(work, "existing-dir")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	runMvdOk(t, src, dst)
	finalPath := filepath.Join(dst, "mysrc")
	assertHistoryChain(t, src, src, finalPath)
	assertFileExists(t, filepath.Join(finalPath, "f.txt"))
}

func testMultiStepMove(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	d2 := filepath.Join(work, "d2")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))
	assertNoErr(t, os.MkdirAll(d2, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")
	assertFileExists(t, p1)

	runMvdOk(t, p1, d2)
	p2 := filepath.Join(d2, "src")
	assertFileExists(t, p2)

	assertHistoryChain(t, src, src, p1, p2)
}

func testMoveAsRootPath(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	d2 := filepath.Join(work, "d2")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))
	assertNoErr(t, os.MkdirAll(d2, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")

	runMvdOk(t, src, d2)
	p2 := filepath.Join(d2, "src")

	assertHistoryChain(t, src, src, p1, p2)
	assertFileExists(t, p2)
	assertFileNotExists(t, p1)
}

func testMoveByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst := filepath.Join(work, "archive")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", src)
	assertHistoryChain(t, src, src)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	runMvdOk(t, "kool", dst)
	p := filepath.Join(dst, "kool")
	assertHistoryChain(t, src, src, p)
	assertFileExists(t, p)
}

func testMoveByAlias(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst1 := filepath.Join(work, "scratch")
	dst2 := filepath.Join(work, "final")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst1, 0755))
	assertNoErr(t, os.MkdirAll(dst2, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, src, dst1)
	p1 := filepath.Join(dst1, "kool")

	runMvdOk(t, "--add-alias", "kk", "kool")

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	runMvdOk(t, "kk", dst2)
	p2 := filepath.Join(dst2, "kool")
	assertHistoryChain(t, src, src, p1, p2)
	assertFileExists(t, p2)
}

func testAmbiguousBasename(t *T) {
	work := testDir(t, "work")
	first := filepath.Join(work, "projects", "kool")
	second := filepath.Join(work, "projects", "v2", "kool")
	dst := filepath.Join(work, "dst")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(first, 0755))
	assertNoErr(t, os.MkdirAll(second, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", first)
	runMvdOk(t, "--add", second)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	out := runMvdErr(t, "kool", dst)
	assertContains(t, out, "ambiguous root basename kool")
}
