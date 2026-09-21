package githook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnstageFilesEmptyPathsIsNoop(t *testing.T) {
	// Outside any repo: a non-noop implementation would try to run git and
	// fail because there is no repository here.
	dir := t.TempDir()
	t.Chdir(dir)
	if err := UnstageFiles(); err != nil {
		t.Fatalf("expected no-op for empty paths, got %v", err)
	}
}

func TestUnstageFilesOnFreshRepoRemovesFromIndex(t *testing.T) {
	repo := initUnstageRepo(t)
	t.Chdir(repo)

	writeUnstageFile(t, filepath.Join(repo, "keep.txt"), "hello\n")
	writeUnstageFile(t, filepath.Join(repo, "drop.bin"), "\x00\x01\x02\x03")
	mustUnstageRun(t, repo, "git", "add", "keep.txt", "drop.bin")

	if err := UnstageFiles("drop.bin"); err != nil {
		t.Fatalf("UnstageFiles: %v", err)
	}

	staged := unstageStagedFiles(t)
	if len(staged) != 1 || staged[0] != "keep.txt" {
		t.Fatalf("expected only keep.txt staged, got %v", staged)
	}
	if _, err := os.Stat(filepath.Join(repo, "drop.bin")); err != nil {
		t.Fatalf("drop.bin must remain on disk: %v", err)
	}
}

func TestUnstageFilesWithCommitRestoresToHead(t *testing.T) {
	repo := initUnstageRepo(t)
	t.Chdir(repo)

	writeUnstageFile(t, filepath.Join(repo, "tracked.txt"), "v1\n")
	mustUnstageRun(t, repo, "git", "add", "tracked.txt")
	mustUnstageRun(t, repo, "git", "commit", "-m", "init")

	// Stage a modification of the tracked file.
	writeUnstageFile(t, filepath.Join(repo, "tracked.txt"), "v2\n")
	mustUnstageRun(t, repo, "git", "add", "tracked.txt")

	if err := UnstageFiles("tracked.txt"); err != nil {
		t.Fatalf("UnstageFiles: %v", err)
	}

	// restore --staged semantics: the index entry resets to HEAD, so the
	// file must remain tracked. git rm --cached here would untrack it and
	// drop the file from the repository on the next commit.
	if !unstageTracked(t, "tracked.txt") {
		t.Fatal("tracked.txt must remain tracked after unstage")
	}
	if staged := unstageStagedFiles(t); len(staged) != 0 {
		t.Fatalf("expected no staged files after unstage, got %v", staged)
	}
	// The working tree keeps the modification.
	data, err := os.ReadFile(filepath.Join(repo, "tracked.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "v2\n" {
		t.Fatalf("working tree content changed: %q", data)
	}
}

func TestUnstageFilesMultiplePathsFreshRepo(t *testing.T) {
	repo := initUnstageRepo(t)
	t.Chdir(repo)

	for _, name := range []string{"a.bin", "b.bin", "keep.txt"} {
		writeUnstageFile(t, filepath.Join(repo, name), "\x00\x01"+name)
	}
	mustUnstageRun(t, repo, "git", "add", "a.bin", "b.bin", "keep.txt")

	if err := UnstageFiles("a.bin", "b.bin"); err != nil {
		t.Fatalf("UnstageFiles: %v", err)
	}

	staged := unstageStagedFiles(t)
	if len(staged) != 1 || staged[0] != "keep.txt" {
		t.Fatalf("expected only keep.txt staged, got %v", staged)
	}
}

func initUnstageRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustUnstageRun(t, repo, "git", "init")
	mustUnstageRun(t, repo, "git", "config", "user.email", "test@example.com")
	mustUnstageRun(t, repo, "git", "config", "user.name", "Test User")
	return repo
}

func writeUnstageFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustUnstageRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}

func unstageStagedFiles(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git diff --cached --name-only: %v", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func unstageTracked(t *testing.T, name string) bool {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--", name)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	return strings.TrimSpace(string(out)) != ""
}
