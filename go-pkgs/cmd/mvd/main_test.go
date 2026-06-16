package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	llsconfig "github.com/xhd2015/lls/config"
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[dir]
	if len(locs) != 1 || locs[0].Path != dir {
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(hist) != 1 {
		t.Fatalf("expected duplicate add to leave history unchanged, got %#v", hist)
	}
	locs := hist[src]
	if len(locs) != 2 || locs[0].Path != src || locs[1].Path != movedPath {
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
	if err := run([]string{"--rm", dir}); err != nil {
		t.Fatalf("run --rm: %v", err)
	}

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[dir]; ok {
		t.Fatalf("expected %s to be removed from history, got %#v", dir, hist)
	}
}

func TestRunRemoveLongAliasDeletesExactSingleEntry(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(work, "tracked")
	mustMkdirAll(t, dir)

	if err := cmdAdd(dir); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}
	if err := run([]string{"--remove", dir}); err != nil {
		t.Fatalf("run --remove: %v", err)
	}

	hist, _, err := loadHistory()
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 2 || locs[0].Path != src || locs[1].Path != movedPath {
		t.Fatalf("expected move history to remain unchanged, got %#v", locs)
	}
}

func TestRemoveBareNameUsesWhichDefault(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	projectRoot := filepath.Join(work, "projects")
	t.Setenv("X", projectRoot)

	configFile, err := llsconfig.DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X"]}`), 0644); err != nil {
		t.Fatalf("write lls config: %v", err)
	}

	basenameProject := filepath.Join(projectRoot, "opencode")
	aliasProject := filepath.Join(projectRoot, "opencode-latest")
	cwd := filepath.Join(work, "cwd")
	mustMkdirAll(t, basenameProject)
	mustMkdirAll(t, aliasProject)
	mustMkdirAll(t, cwd)

	if err := cmdAdd(basenameProject); err != nil {
		t.Fatalf("cmdAdd basenameProject: %v", err)
	}
	if err := cmdAdd(aliasProject); err != nil {
		t.Fatalf("cmdAdd aliasProject: %v", err)
	}
	if err := cmdAddAlias("opencode", "$X/opencode-latest"); err != nil {
		t.Fatalf("cmdAddAlias: %v", err)
	}

	t.Chdir(cwd)
	output := captureStdout(t, func() {
		if err := run([]string{"--rm", "opencode"}); err != nil {
			t.Fatalf("run --rm opencode: %v", err)
		}
	})
	if !strings.Contains(output, "removed: $X/opencode") {
		t.Fatalf("expected remove output for basename project, got %q", output)
	}

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[basenameProject]; ok {
		t.Fatalf("expected basename project %s to be removed, got %#v", basenameProject, hist)
	}
	if _, ok := hist[aliasProject]; !ok {
		t.Fatalf("expected alias project %s to remain, got %#v", aliasProject, hist)
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
	if !strings.Contains(err.Error(), "has movement history:\n  use `mvd --rm -f") {
		t.Fatalf("expected wrapped force hint in error, got %q", err.Error())
	}

	hist, _, loadErr := loadHistory()
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
		if err := run([]string{"--rm", "-f", src}); err != nil {
			t.Fatalf("run --rm -f: %v", err)
		}
	})
	if !strings.Contains(output, "will clear") {
		t.Fatalf("expected history-clearing hint, got %q", output)
	}

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[src]; ok {
		t.Fatalf("expected %s history to be removed, got %#v", src, hist[src])
	}
}

func TestForceRequiresRemoveFlag(t *testing.T) {
	err := run([]string{"-f", "src", "dst"})
	if err == nil {
		t.Fatalf("expected force without remove to fail")
	}
	if !strings.Contains(err.Error(), "requires --rm") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunRejectsMultipleModeFlags(t *testing.T) {
	modes := []string{"--rm", "--add", "--rebase", "--list", "--back", "--clear", "--print", "--worktree"}
	for i, first := range modes {
		for _, second := range modes[i+1:] {
			t.Run(first+"_"+second, func(t *testing.T) {
				err := run([]string{first, second, "src", "dst"})
				if err == nil {
					t.Fatalf("expected %s and %s to conflict", first, second)
				}
				if !strings.Contains(err.Error(), "at most one of --rm, --add, --add-alias, --rebase, --list, --which, --back, --clear, --print") {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		}
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[src]; ok {
		t.Fatalf("expected old base %s to be removed", src)
	}
	locs := hist[newBase]
	expectedCurrent := filepath.Join(dst, "dir")
	if len(locs) != 3 || locs[0].Path != newBase || locs[1].Path != src || locs[2].Path != expectedCurrent {
		t.Fatalf("expected rebased history [%s %s %s], got %#v", newBase, src, expectedCurrent, locs)
	}
}

func TestRunRebaseFlagChangesEntryBase(t *testing.T) {
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
	if err := run([]string{"--rebase", src, newBase}); err != nil {
		t.Fatalf("run --rebase: %v", err)
	}

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if _, ok := hist[src]; ok {
		t.Fatalf("expected old base %s to be removed", src)
	}
	if locs := hist[newBase]; len(locs) == 0 || locs[0].Path != newBase {
		t.Fatalf("expected rebased history under %s, got %#v", newBase, locs)
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	expectedCurrent := filepath.Join(dst2, "dir")
	locs := hist[newBase]
	if len(locs) != 4 || locs[0].Path != newBase || locs[1].Path != src || locs[2].Path != historyPath || locs[3].Path != expectedCurrent {
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 3 || locs[0].Path != src || locs[1].Path != firstPath || locs[2].Path != secondPath {
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 3 || locs[0].Path != src || locs[1].Path != firstPath || locs[2].Path != secondPath {
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	trackedLocs := hist[tracked]
	if len(trackedLocs) != 2 || trackedLocs[0].Path != tracked || trackedLocs[1].Path != trackedCurrent {
		t.Fatalf("expected tracked history unchanged, got %#v", trackedLocs)
	}
	localLocs := hist[local]
	if len(localLocs) != 2 || localLocs[0].Path != local || localLocs[1].Path != localMoved {
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

func TestRunAddAliasStoresAliasForProjectBasename(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "projects", "kool")
	mustMkdirAll(t, src)

	if err := cmdAdd(src); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}
	if err := run([]string{"--add-alias", "kk", "kool"}); err != nil {
		t.Fatalf("run --add-alias: %v", err)
	}

	_, aliases, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if aliases["kk"] != src {
		t.Fatalf("expected alias kk to point at %s, got %#v", src, aliases)
	}
}

func TestMoveAcceptsAliasAfterBasenameMiss(t *testing.T) {
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
	if err := cmdAddAlias("kk", "kool"); err != nil {
		t.Fatalf("cmdAddAlias: %v", err)
	}

	t.Chdir(cwd)
	if err := cmdMove("kk", dst2); err != nil {
		t.Fatalf("second cmdMove with alias: %v", err)
	}

	firstPath := filepath.Join(dst1, "kool")
	secondPath := filepath.Join(dst2, "kool")
	if pathExists(firstPath) {
		t.Fatalf("expected %s to be moved onward", firstPath)
	}
	if !pathExists(secondPath) {
		t.Fatalf("expected %s to exist after alias move", secondPath)
	}

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 3 || locs[0].Path != src || locs[1].Path != firstPath || locs[2].Path != secondPath {
		t.Fatalf("expected move history [%s %s %s], got %#v", src, firstPath, secondPath, locs)
	}
}

func TestMovePrefersBasenameOverAlias(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	aliasProject := filepath.Join(work, "projects", "aliased")
	basenameProject := filepath.Join(work, "projects", "kk")
	aliasDst := filepath.Join(work, "alias-dst")
	basenameDst := filepath.Join(work, "basename-dst")
	cwd := filepath.Join(work, "cwd")

	mustMkdirAll(t, aliasProject)
	mustMkdirAll(t, basenameProject)
	mustMkdirAll(t, aliasDst)
	mustMkdirAll(t, basenameDst)
	mustMkdirAll(t, cwd)

	if err := cmdMove(aliasProject, aliasDst); err != nil {
		t.Fatalf("alias project cmdMove: %v", err)
	}
	if err := cmdMove(basenameProject, basenameDst); err != nil {
		t.Fatalf("basename project cmdMove: %v", err)
	}
	if err := cmdAddAlias("kk", "aliased"); err != nil {
		t.Fatalf("cmdAddAlias: %v", err)
	}

	finalDst := filepath.Join(work, "final")
	mustMkdirAll(t, finalDst)
	t.Chdir(cwd)
	if err := cmdMove("kk", finalDst); err != nil {
		t.Fatalf("cmdMove should prefer basename over alias: %v", err)
	}

	if !pathExists(filepath.Join(aliasDst, "aliased")) {
		t.Fatalf("expected alias target to remain in place")
	}
	if !pathExists(filepath.Join(finalDst, "kk")) {
		t.Fatalf("expected basename target to move")
	}
}

func TestMovePrefersLocalPathOverAlias(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	aliasProject := filepath.Join(work, "projects", "aliased")
	aliasDst := filepath.Join(work, "alias-dst")
	localDst := filepath.Join(work, "local-dst")
	cwd := filepath.Join(work, "cwd")

	mustMkdirAll(t, aliasProject)
	mustMkdirAll(t, aliasDst)
	mustMkdirAll(t, localDst)
	mustMkdirAll(t, cwd)

	if err := cmdMove(aliasProject, aliasDst); err != nil {
		t.Fatalf("alias project cmdMove: %v", err)
	}
	if err := cmdAddAlias("kk", "aliased"); err != nil {
		t.Fatalf("cmdAddAlias: %v", err)
	}

	t.Chdir(cwd)
	local := filepath.Join(cwd, "kk")
	if err := os.WriteFile(local, []byte("local"), 0644); err != nil {
		t.Fatalf("write local src: %v", err)
	}
	if err := cmdMove("kk", localDst); err != nil {
		t.Fatalf("cmdMove should prefer local path over alias: %v", err)
	}

	if !pathExists(filepath.Join(localDst, "kk")) {
		t.Fatalf("expected local source to move")
	}
	if !pathExists(filepath.Join(aliasDst, "aliased")) {
		t.Fatalf("expected alias target to remain in place")
	}
}

func TestListAllShortensPathsWithLLSConfig(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	projectRoot := filepath.Join(work, "projects")
	t.Setenv("X", projectRoot)

	configFile, err := llsconfig.DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X"]}`), 0644); err != nil {
		t.Fatalf("write lls config: %v", err)
	}

	src := filepath.Join(projectRoot, "kool")
	dst := filepath.Join(projectRoot, "archive")
	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}

	output := captureStdout(t, func() {
		if err := cmdListAll(); err != nil {
			t.Fatalf("cmdListAll: %v", err)
		}
	})
	if !strings.Contains(output, "$X/kool") {
		t.Fatalf("expected shortened root path, got %q", output)
	}
	if !strings.Contains(output, "$X/archive/kool") {
		t.Fatalf("expected shortened latest path, got %q", output)
	}
	if strings.Contains(output, projectRoot) {
		t.Fatalf("expected output to hide long project root %s, got %q", projectRoot, output)
	}
}

func TestListAllShowsAliasesOnRootLine(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	projectRoot := filepath.Join(work, "projects")
	t.Setenv("X", projectRoot)

	configFile, err := llsconfig.DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X"]}`), 0644); err != nil {
		t.Fatalf("write lls config: %v", err)
	}

	src := filepath.Join(projectRoot, "opencode-latest")
	dst := filepath.Join(projectRoot, "archive")
	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := cmdMove(src, dst); err != nil {
		t.Fatalf("cmdMove: %v", err)
	}
	if err := cmdAddAlias("opencode", "$X/opencode-latest"); err != nil {
		t.Fatalf("cmdAddAlias: %v", err)
	}

	output := captureStdout(t, func() {
		if err := cmdListAll(); err != nil {
			t.Fatalf("cmdListAll: %v", err)
		}
	})

	if !strings.Contains(output, "$X/opencode-latest (aliases: opencode)") {
		t.Fatalf("expected list output to include alias on root line, got %q", output)
	}
}

func TestWhichShowsAllMatchesInMoveResolutionOrder(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	projectRoot := filepath.Join(work, "projects")
	t.Setenv("X", projectRoot)

	configFile, err := llsconfig.DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X"]}`), 0644); err != nil {
		t.Fatalf("write lls config: %v", err)
	}

	localDir := filepath.Join(work, "cwd")
	local := filepath.Join(localDir, "opencode")
	basenameProject := filepath.Join(projectRoot, "opencode")
	aliasProject := filepath.Join(projectRoot, "opencode-latest")
	mustMkdirAll(t, localDir)
	mustMkdirAll(t, basenameProject)
	mustMkdirAll(t, aliasProject)
	t.Chdir(localDir)

	if err := cmdAdd(basenameProject); err != nil {
		t.Fatalf("cmdAdd basenameProject: %v", err)
	}
	if err := cmdAdd(aliasProject); err != nil {
		t.Fatalf("cmdAdd aliasProject: %v", err)
	}
	if err := cmdAddAlias("opencode", "$X/opencode-latest"); err != nil {
		t.Fatalf("cmdAddAlias: %v", err)
	}
	if err := os.WriteFile(local, []byte("local"), 0644); err != nil {
		t.Fatalf("write local: %v", err)
	}

	output := captureStdout(t, func() {
		if err := run([]string{"--which", "opencode"}); err != nil {
			t.Fatalf("run --which: %v", err)
		}
	})

	want := strings.Join([]string{
		local + " (local)",
		"$X/opencode (project basename)",
		"$X/opencode-latest (alias: opencode)",
		"",
	}, "\n")
	if output != want {
		t.Fatalf("expected %q, got %q", want, output)
	}
}

func TestPrintShowsShortAndFullPathForUniqueBasename(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	projectRoot := filepath.Join(work, "projects")
	t.Setenv("X", projectRoot)

	configFile, err := llsconfig.DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X"]}`), 0644); err != nil {
		t.Fatalf("write lls config: %v", err)
	}

	src := filepath.Join(projectRoot, "kool")
	cwd := filepath.Join(work, "cwd")
	mustMkdirAll(t, src)
	mustMkdirAll(t, cwd)
	t.Chdir(cwd)

	if err := cmdAdd(src); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}

	output := captureStdout(t, func() {
		if err := run([]string{"--print", "kool"}); err != nil {
			t.Fatalf("run --print: %v", err)
		}
	})

	want := "$X/kool -> " + src + "\n"
	if output != want {
		t.Fatalf("expected %q, got %q", want, output)
	}
}

func TestPrintShortFlagAlias(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "kool")
	mustMkdirAll(t, src)

	output := captureStdout(t, func() {
		if err := run([]string{"-p", src}); err != nil {
			t.Fatalf("run -p: %v", err)
		}
	})

	want := src + " -> " + src + "\n"
	if output != want {
		t.Fatalf("expected %q, got %q", want, output)
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 1 || locs[0].Path != src {
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 1 || locs[0].Path != src {
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

	hist, _, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	locs := hist[src]
	if len(locs) != 1 || locs[0].Path != src {
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

func resetDisplayConfig() {
	displayConfigOnce = sync.Once{}
	displayConfig = nil
}

func TestBuildProjectListEmptyHistory(t *testing.T) {
	hist := History{}
	aliases := map[string]string{}

	entries := buildProjectList(hist, aliases)
	if len(entries) != 0 {
		t.Fatalf("expected empty list, got %#v", entries)
	}
}

func TestBuildProjectListBasicEntries(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	hist := History{
		"/path/to/foo": {{Path: "/path/to/foo"}, {Path: "/path/to/foo-new"}},
		"/path/to/bar": {{Path: "/path/to/bar"}},
	}
	aliases := map[string]string{}

	entries := buildProjectList(hist, aliases)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// display includes (dead) markers because test paths don't exist on disk
	if entries[0].full != "/path/to/bar" || !strings.HasPrefix(entries[0].display, "/path/to/bar") {
		t.Fatalf("expected first entry bar, got %#v", entries[0])
	}
	if entries[1].full != "/path/to/foo-new" || !strings.HasPrefix(entries[1].display, "/path/to/foo-new") {
		t.Fatalf("expected second entry foo-new, got %#v", entries[1])
	}
}

func TestBuildProjectListSkipsEmptyLocations(t *testing.T) {
	hist := History{
		"/path/to/empty": {},
		"/path/to/real": {{Path: "/path/to/real"}},
	}

	entries := buildProjectList(hist, nil)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].full != "/path/to/real" {
		t.Fatalf("expected real entry, got %#v", entries[0])
	}
}

func TestBuildProjectListAnnotatesAliases(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	hist := History{
		"/path/to/foo": {{Path: "/path/to/foo"}, {Path: "/path/to/foo-latest"}},
	}
	aliases := map[string]string{
		"f":  "/path/to/foo",
		"ff": "/path/to/foo",
	}

	entries := buildProjectList(hist, aliases)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// aliases now appear on the latest (non-root) entry for plain moves;
	// display also includes (dead) marker since test paths don't exist on disk
	if !strings.Contains(entries[0].display, "aliases: f, ff") {
		t.Fatalf("expected alias annotation, got %q", entries[0].display)
	}
	if !strings.Contains(entries[0].display, "ff") {
		t.Fatalf("expected alias ff in display, got %q", entries[0].display)
	}
	if !strings.Contains(entries[0].display, "f,") && !strings.Contains(entries[0].display, "f)") {
		t.Fatalf("expected alias f in display, got %q", entries[0].display)
	}
	if entries[0].full != "/path/to/foo-latest" {
		t.Fatalf("expected full path to be latest, got %q", entries[0].full)
	}
}

func TestBuildProjectListDeduplicatesByDisplayPath(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	hist := History{
		"/path/to/foo": {{Path: "/path/to/foo"}, {Path: "/path/to/shared"}},
		"/path/to/bar": {{Path: "/path/to/bar"}, {Path: "/path/to/shared"}},
	}

	entries := buildProjectList(hist, nil)
	if len(entries) != 1 {
		t.Fatalf("expected 1 deduplicated entry, got %d: %#v", len(entries), entries)
	}
	if entries[0].full != "/path/to/shared" {
		t.Fatalf("expected /path/to/shared, got %q", entries[0].full)
	}
}

func TestRunNoArgsNonTTYShowsHelp(t *testing.T) {
	output := captureStdout(t, func() {
		if err := run([]string{}); err != nil {
			t.Fatalf("run(): %v", err)
		}
	})
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestRunPrintNoArgsNonTTYReturnsError(t *testing.T) {
	err := run([]string{"--print"})
	if err == nil {
		t.Fatal("expected error for --print with no args on non-TTY")
	}
	if !strings.Contains(err.Error(), "usage: mvd --print") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunVscodeNoArgsNonTTYReturnsError(t *testing.T) {
	err := run([]string{"--vscode"})
	if err == nil {
		t.Fatal("expected error for --vscode with no args on non-TTY")
	}
	if !strings.Contains(err.Error(), "usage: mvd --vscode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCdNoArgsNonTTYReturnsError(t *testing.T) {
	err := run([]string{"--cd"})
	if err == nil {
		t.Fatal("expected error for --cd with no args on non-TTY")
	}
	if !strings.Contains(err.Error(), "usage: mvd --cd") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildProjectListAliasesUseRootKey(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	hist := History{
		"/path/to/root": {{Path: "/path/to/root"}, {Path: "/path/to/moved"}},
	}
	aliases := map[string]string{
		"alias": "/path/to/moved",
	}

	entries := buildProjectList(hist, aliases)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if strings.Contains(entries[0].display, "(aliases:") {
		t.Fatalf("alias should only annotate root key, got %q", entries[0].display)
	}
}

func TestBuildProjectListAliasOnProperRoot(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	hist := History{
		"/path/to/root": {{Path: "/path/to/root"}, {Path: "/path/to/latest"}},
	}
	aliases := map[string]string{
		"alias": "/path/to/root",
	}

	entries := buildProjectList(hist, aliases)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// aliases now appear on the latest entry for plain moves (not just root);
	// display also includes (dead) marker since test paths don't exist on disk
	if !strings.Contains(entries[0].display, "aliases: alias") {
		t.Fatalf("expected alias annotation, got %q", entries[0].display)
	}
	if entries[0].full != "/path/to/latest" {
		t.Fatalf("expected full path to be latest, got %q", entries[0].full)
	}
}

func TestBuildProjectListWithEnvCollapsing(t *testing.T) {
	resetDisplayConfig()
	t.Cleanup(resetDisplayConfig)

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	projectRoot := filepath.Join(work, "projects")
	t.Setenv("X", projectRoot)

	configFile, err := llsconfig.DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X"]}`), 0644); err != nil {
		t.Fatalf("write lls config: %v", err)
	}

	hist := History{
		filepath.Join(projectRoot, "foo"): {{Path: filepath.Join(projectRoot, "foo")}, {Path: filepath.Join(projectRoot, "foo-new")}},
	}

	entries := buildProjectList(hist, nil)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !strings.Contains(entries[0].display, "$X/foo-new") {
		t.Fatalf("expected env-collapsed display path, got %q", entries[0].display)
	}
}

func TestRunNoModeNoArgsPipedStdinShowsHelp(t *testing.T) {
	_, w, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = w
	defer func() {
		os.Stdin = oldStdin
		w.Close()
	}()

	output := captureStdout(t, func() {
		err := run([]string{})
		if err != nil {
			t.Fatalf("run(): %v", err)
		}
	})

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output with piped stdin, got %q", output)
	}
}

func TestRunStillAllowsSingleArgMove(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	mustMkdirAll(t, src)
	mustMkdirAll(t, dst)

	if err := run([]string{src, dst}); err != nil {
		t.Fatalf("run move: %v", err)
	}

	movedPath := filepath.Join(dst, "src")
	if !pathExists(movedPath) {
		t.Fatalf("expected %s to exist", movedPath)
	}
}
