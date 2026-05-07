package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAddFlagStoresSingleEntry(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(work, "tracked")
	mustMkdirAll(t, dir)

	if err := run([]string{"--add", dir}); err != nil {
		t.Fatalf("run --add: %v", err)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[dir]
	if len(locs) != 1 || locs[0] != dir {
		t.Fatalf("expected single record entry for %s, got %#v", dir, locs)
	}
}

func TestRunMoveCanUseAddAsSource(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	cwd := filepath.Join(work, "cwd")
	dst := filepath.Join(work, "dst")
	mustMkdirAll(t, cwd)
	mustMkdirAll(t, dst)
	t.Chdir(cwd)

	src := filepath.Join(cwd, "add")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := run([]string{"add", dst}); err != nil {
		t.Fatalf("run move with add source: %v", err)
	}

	movedPath := filepath.Join(dst, "add")
	if !pathExists(movedPath) {
		t.Fatalf("expected %s to exist after move", movedPath)
	}
	if pathExists(src) {
		t.Fatalf("expected %s to be moved", src)
	}
}

func TestAddExistingHistoryPathDoesNothing(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	movedPath := filepath.Join(dst, "dir")
	output := captureStdout(t, func() {
		if err := cmdAdd(movedPath); err != nil {
			t.Fatalf("cmdAdd: %v", err)
		}
	})
	if !bytes.Contains([]byte(output), []byte("already recorded")) {
		t.Fatalf("expected duplicate hint, got %q", output)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(hist) != 1 {
		t.Fatalf("expected duplicate add to leave history unchanged, got %#v", hist)
	}
	locs := hist[src]
	if len(locs) != 2 || locs[0] != src || locs[1] != movedPath {
		t.Fatalf("expected original move history to remain unchanged, got %#v", locs)
	}
}

func TestRunRemoveDeletesExactSingleEntry(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(work, "tracked")
	mustMkdirAll(t, dir)

	if err := cmdAdd(dir); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}
	if err := run([]string{"rm", dir}); err != nil {
		t.Fatalf("run rm: %v", err)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[dir]; ok {
		t.Fatalf("expected %s to be removed from history, got %#v", dir, hist)
	}
}

func TestRemoveDoesNotMatchHistoryOnlyPath(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	movedPath := filepath.Join(dst, "dir")
	output := captureStdout(t, func() {
		if err := cmdRemove(movedPath, false); err != nil {
			t.Fatalf("cmdRemove: %v", err)
		}
	})
	if !strings.Contains(output, "no recorded entry") {
		t.Fatalf("expected missing-entry hint, got %q", output)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 2 || locs[0] != src || locs[1] != movedPath {
		t.Fatalf("expected move history to remain unchanged, got %#v", locs)
	}
}

func TestRemoveWithHistoryRequiresForce(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	err := cmdRemove(src, false)
	if err == nil {
		t.Fatalf("expected cmdRemove to fail without force")
	}
	if !strings.Contains(err.Error(), "has movement history:\n  use `mvd rm -f") {
		t.Fatalf("expected wrapped force hint in error, got %q", err.Error())
	}

	hist, loadErr := loadHistory()
	if loadErr != nil {
		t.Fatalf("loadHistory: %v", loadErr)
	}
	if _, ok := hist[src]; !ok {
		t.Fatalf("expected history for %s to remain after failed remove", src)
	}
}

func TestRemoveWithForceClearsHistory(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	output := captureStdout(t, func() {
		if err := run([]string{"rm", "-f", src}); err != nil {
			t.Fatalf("run rm -f: %v", err)
		}
	})
	if !strings.Contains(output, "will clear") {
		t.Fatalf("expected history-clearing hint, got %q", output)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[src]; ok {
		t.Fatalf("expected %s history to be removed, got %#v", src, hist[src])
	}
}

func TestRebaseChangesEntryBaseFromKeyMatch(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")
	newBase := filepath.Join(work, "rebased", "dir")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}
	if err := cmdRebase(src, newBase); err != nil {
		t.Fatalf("cmdRebase: %v", err)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[src]; ok {
		t.Fatalf("expected old base %s to be removed", src)
	}
	locs := hist[newBase]
	expectedCurrent := filepath.Join(dst, "dir")
	if len(locs) != 3 || locs[0] != newBase || locs[1] != src || locs[2] != expectedCurrent {
		t.Fatalf("expected rebased history [%s %s %s], got %#v", newBase, src, expectedCurrent, locs)
	}
}

func TestRebaseFindsEntryByHistoryMatch(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst1 := filepath.Join(work, "b")
	dst2 := filepath.Join(work, "c")
	newBase := filepath.Join(work, "rebased", "dir")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst1)
	mustMkdirAll(t, dst2)

	if err := cmdMove(src, dst1); err != nil {
		t.Fatalf("first cmdMove: %v", err)
	}
	if err := cmdMove(filepath.Join(dst1, "dir"), dst2); err != nil {
		t.Fatalf("second cmdMove: %v", err)
	}

	historyPath := filepath.Join(dst1, "dir")
	if err := cmdRebase(historyPath, newBase); err != nil {
		t.Fatalf("cmdRebase: %v", err)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	expectedCurrent := filepath.Join(dst2, "dir")
	locs := hist[newBase]
	if len(locs) != 4 || locs[0] != newBase || locs[1] != src || locs[2] != historyPath || locs[3] != expectedCurrent {
		t.Fatalf("expected rebased history [%s %s %s %s], got %#v", newBase, src, historyPath, expectedCurrent, locs)
	}
}

func TestRebaseRejectsNewBaseOwnedByAnotherEntry(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	first := filepath.Join(work, "first")
	second := filepath.Join(work, "second")

	mustMkdirAll(t, first)
	mustMkdirAll(t, second)

	if err := cmdAdd(first); err != nil {
		t.Fatalf("cmdAdd first: %v", err)
	}
	if err := cmdAdd(second); err != nil {
		t.Fatalf("cmdAdd second: %v", err)
	}

	err := cmdRebase(first, second)
	if err == nil {
		t.Fatalf("expected cmdRebase to reject duplicate new base")
	}
	if !strings.Contains(err.Error(), "already recorded in another entry") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMoveAcceptsOriginalRootPath(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst1 := filepath.Join(work, "b")
	dst2 := filepath.Join(work, "c")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst1)
	mustMkdirAll(t, dst2)

	if err := cmdMove(src, dst1); err != nil {
		t.Fatalf("first cmdMove: %v", err)
	}
	if err := cmdMove(src, dst2); err != nil {
		t.Fatalf("second cmdMove with root path: %v", err)
	}

	firstPath := filepath.Join(dst1, "dir")
	secondPath := filepath.Join(dst2, "dir")
	if pathExists(firstPath) {
		t.Fatalf("expected %s to be moved onward", firstPath)
	}
	if !pathExists(secondPath) {
		t.Fatalf("expected %s to exist after root-path move", secondPath)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 3 || locs[0] != src || locs[1] != firstPath || locs[2] != secondPath {
		t.Fatalf("expected move history [%s %s %s], got %#v", src, firstPath, secondPath, locs)
	}
}

func TestMoveAcceptsUniqueOriginalBasename(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "projects", "kool")
	dst1 := filepath.Join(work, "scratch")
	dst2 := filepath.Join(work, "final")
	cwd := filepath.Join(work, "cwd")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst1)
	mustMkdirAll(t, dst2)
	mustMkdirAll(t, cwd)

	if err := cmdMove(src, dst1); err != nil {
		t.Fatalf("first cmdMove: %v", err)
	}

	t.Chdir(cwd)
	if err := cmdMove("kool", dst2); err != nil {
		t.Fatalf("second cmdMove with unique root basename: %v", err)
	}

	firstPath := filepath.Join(dst1, "kool")
	secondPath := filepath.Join(dst2, "kool")
	if pathExists(firstPath) {
		t.Fatalf("expected %s to be moved onward", firstPath)
	}
	if !pathExists(secondPath) {
		t.Fatalf("expected %s to exist after basename move", secondPath)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 3 || locs[0] != src || locs[1] != firstPath || locs[2] != secondPath {
		t.Fatalf("expected move history [%s %s %s], got %#v", src, firstPath, secondPath, locs)
	}
}

func TestMoveDoesNotUseBasenameShortcutWhenLocalPathExists(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	tracked := filepath.Join(work, "projects", "kool")
	trackedDst := filepath.Join(work, "scratch")
	localDst := filepath.Join(work, "local-dst")
	cwd := filepath.Join(work, "cwd")

	mustMkdirAll(t, tracked)
	mustMkdirAll(t, trackedDst)
	mustMkdirAll(t, localDst)
	mustMkdirAll(t, cwd)

	if err := cmdMove(tracked, trackedDst); err != nil {
		t.Fatalf("tracked cmdMove: %v", err)
	}

	t.Chdir(cwd)
	local := filepath.Join(cwd, "kool")
	if err := os.WriteFile(local, []byte("local"), 0644); err != nil {
		t.Fatalf("write local src: %v", err)
	}

	if err := cmdMove("kool", localDst); err != nil {
		t.Fatalf("cmdMove should use local source: %v", err)
	}

	localMoved := filepath.Join(localDst, "kool")
	trackedCurrent := filepath.Join(trackedDst, "kool")
	if !pathExists(localMoved) {
		t.Fatalf("expected local source to move to %s", localMoved)
	}
	if !pathExists(trackedCurrent) {
		t.Fatalf("expected tracked basename shortcut target %s to remain in place", trackedCurrent)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	trackedLocs := hist[tracked]
	if len(trackedLocs) != 2 || trackedLocs[0] != tracked || trackedLocs[1] != trackedCurrent {
		t.Fatalf("expected tracked history unchanged, got %#v", trackedLocs)
	}
	localLocs := hist[local]
	if len(localLocs) != 2 || localLocs[0] != local || localLocs[1] != localMoved {
		t.Fatalf("expected local move history [%s %s], got %#v", local, localMoved, localLocs)
	}
}

func TestMoveRejectsDuplicateOriginalBasename(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	first := filepath.Join(work, "projects", "kool")
	second := filepath.Join(work, "projects", "v2", "kool")
	dst := filepath.Join(work, "dst")
	cwd := filepath.Join(work, "cwd")

	mustMkdirAll(t, first)
	mustMkdirAll(t, second)
	mustMkdirAll(t, dst)
	mustMkdirAll(t, cwd)

	if err := cmdAdd(first); err != nil {
		t.Fatalf("cmdAdd first: %v", err)
	}
	if err := cmdAdd(second); err != nil {
		t.Fatalf("cmdAdd second: %v", err)
	}

	t.Chdir(cwd)
	err := cmdMove("kool", dst)
	if err == nil {
		t.Fatalf("expected duplicate basename to fail")
	}
	if !strings.Contains(err.Error(), "ambiguous root basename kool") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBackKeepsOriginalHistoryEntry(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	movedPath := filepath.Join(dst, "dir")
	if err := cmdBack(movedPath); err != nil {
		t.Fatalf("cmdBack: %v", err)
	}

	if !pathExists(src) {
		t.Fatalf("expected %s to exist after moving back", src)
	}
	if pathExists(movedPath) {
		t.Fatalf("expected %s to be removed after moving back", movedPath)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 1 || locs[0] != src {
		t.Fatalf("expected single original history entry for %s, got %#v", src, locs)
	}
}

func TestBackAcceptsOriginalRootPath(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	movedPath := filepath.Join(dst, "dir")
	if err := cmdBack(src); err != nil {
		t.Fatalf("cmdBack with root path: %v", err)
	}

	if !pathExists(src) {
		t.Fatalf("expected %s to exist after moving back", src)
	}
	if pathExists(movedPath) {
		t.Fatalf("expected %s to be removed after moving back", movedPath)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 1 || locs[0] != src {
		t.Fatalf("expected single original history entry for %s, got %#v", src, locs)
	}
}

func TestBackAcceptsUniqueOriginalBasename(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "projects", "kool")
	dst := filepath.Join(work, "scratch")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	movedPath := filepath.Join(dst, "kool")
	if err := cmdBack("kool"); err != nil {
		t.Fatalf("cmdBack with unique root basename: %v", err)
	}

	if !pathExists(src) {
		t.Fatalf("expected %s to exist after moving back", src)
	}
	if pathExists(movedPath) {
		t.Fatalf("expected %s to be removed after moving back", movedPath)
	}
}

func TestBackDoesNotUseBasenameShortcutWhenLocalPathExists(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "projects", "kool")
	dst := filepath.Join(work, "scratch")
	cwd := filepath.Join(work, "cwd")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)
	mustMkdirAll(t, cwd)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	t.Chdir(cwd)
	localFile := filepath.Join(cwd, "kool")
	if err := os.WriteFile(localFile, []byte("local"), 0644); err != nil {
		t.Fatalf("write local file: %v", err)
	}

	err := cmdBack("kool")
	if err == nil {
		t.Fatalf("expected local path without history to block basename shortcut")
	}
	if !strings.Contains(err.Error(), "no mv history for "+localFile) {
		t.Fatalf("unexpected error: %v", err)
	}

	movedPath := filepath.Join(dst, "kool")
	if !pathExists(movedPath) {
		t.Fatalf("expected basename shortcut target %s to remain in place", movedPath)
	}
	if pathExists(src) {
		t.Fatalf("expected %s not to be restored via basename shortcut", src)
	}
}

func TestBackRejectsDuplicateOriginalBasename(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	first := filepath.Join(work, "projects", "kool")
	second := filepath.Join(work, "projects", "v2", "kool")

	mustMkdirAll(t, first)
	mustMkdirAll(t, second)

	if err := cmdAdd(first); err != nil {
		t.Fatalf("cmdAdd first: %v", err)
	}
	if err := cmdAdd(second); err != nil {
		t.Fatalf("cmdAdd second: %v", err)
	}

	err := cmdBack("kool")
	if err == nil {
		t.Fatalf("expected duplicate basename to fail")
	}
	if !strings.Contains(err.Error(), "ambiguous root basename kool") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBackAtOriginalPositionIsNoOp(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "a", "dir")
	dst := filepath.Join(work, "b")

	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	movedPath := filepath.Join(dst, "dir")
	if err := cmdBack(movedPath); err != nil {
		t.Fatalf("first cmdBack: %v", err)
	}
	if err := cmdBack(src); err != nil {
		t.Fatalf("second cmdBack should be no-op, got: %v", err)
	}

	if !pathExists(src) {
		t.Fatalf("expected %s to remain in place", src)
	}

	hist, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 1 || locs[0] != src {
		t.Fatalf("expected history to remain at original location for %s, got %#v", src, locs)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return string(out)
}
