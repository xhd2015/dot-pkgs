# Scenario

**Feature**: `analyse.Scan` walks a fake HOME; format helpers render entry blocks and summary

```
# scan pipeline
caller Options{Home, OnEntry?} -> Scan -> per-entry EntryResult + ScanSummary

# optional streaming
Scan -> OnEntry(entry) after each top-level child (sorted); error aborts

# format (standalone)
FormatEntryBlock(entry) -> block text
FormatSummaryLines(summary) -> summary lines
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse` is importable.
- Each scan leaf uses an isolated `t.TempDir()` as fake HOME.
- `git` on PATH for `git-dirs` seeding (leaf skips when unavailable).

## Context

- Fixture seeding mirrors `tests/remote-agent-machine-analyse-files` profiles.
- Assertions use structured `EntryResult` / `ScanSummary` fields, not CLI stdout.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse"
)

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, home, rel string, content []byte) {
	t.Helper()
	full := filepath.Join(home, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, content, 0644); err != nil {
		t.Fatal(err)
	}
}

func writeText(t *testing.T, home, rel, content string) {
	t.Helper()
	writeFile(t, home, rel, []byte(content))
}

func gitAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
		return false
	}
	return true
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func gitInitRepo(t *testing.T, dir string) {
	t.Helper()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "branch", "-M", "main")
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("seed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "README.md")
	gitRun(t, dir, "commit", "-m", "Initial commit")
}

func seedCodexTree(t *testing.T, home string) {
	t.Helper()
	writeText(t, home, ".codex/sessions/thread-a/rollout-001.jsonl", `{"type":"session"}`+"\n")
	writeText(t, home, ".codex/sessions/thread-b/rollout-002.jsonl", `{"type":"session"}`+"\n")
	writeText(t, home, ".codex/skills/my-skill/SKILL.md", "# skill\n")
	writeText(t, home, ".codex/cache/warm.dat", "cache-bytes\n")
	writeText(t, home, ".codex/rules/default.md", "# rule\n")
}

func seedPlainDir(t *testing.T, home string) {
	t.Helper()
	writeText(t, home, "plain-dir/sub/nested.txt", "nested\n")
}

func seedNodeModulesEntry(t *testing.T, home string) {
	t.Helper()
	writeText(t, home, "nm-entry/node_modules/pkg/index.js", "module.exports = 1;\n")
	writeText(t, home, "nm-entry/src/deep/node_modules/nested/index.js", "module.exports = 2;\n")
	writeText(t, home, "nm-entry/src/app.js", "console.log('app');\n")
}

func seedGitDirsEntries(t *testing.T, home string) {
	t.Helper()
	withGit := filepath.Join(home, "with-git")
	mkdirAll(t, withGit)
	gitInitRepo(t, withGit)
	seedPlainDir(t, home)
}

func seedEntryOrderEntries(t *testing.T, home string) {
	t.Helper()
	seedCodexTree(t, home)
	writeText(t, home, "aaa-first/alpha.txt", "a\n")
	writeText(t, home, "mmm-mid/middle.txt", "m\n")
	writeText(t, home, "notes.txt", "line one\nline two\n")
	writeText(t, home, "zzz-last/omega.txt", "z\n")
}

func seedFileLinesEntries(t *testing.T, home string) {
	t.Helper()
	notes, err := os.ReadFile("testdata/notes.txt")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile("testdata/binary.dat")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, home, "notes.txt", notes)
	writeFile(t, home, "binary.dat", binary)
}

func seedBasicEntries(t *testing.T, home string) {
	t.Helper()
	seedPlainDir(t, home)
	writeText(t, home, "notes.txt", "alpha\nbeta\n")
}

func seedOnEntryEntries(t *testing.T, home string) {
	t.Helper()
	writeText(t, home, "aaa-first/alpha.txt", "a\n")
	writeText(t, home, "mmm-mid/middle.txt", "m\n")
	writeText(t, home, "zzz-last/omega.txt", "z\n")
}

func seedHome(t *testing.T, home, profile string) {
	t.Helper()
	switch profile {
	case "basic":
		seedBasicEntries(t, home)
	case "codex":
		seedCodexTree(t, home)
		seedPlainDir(t, home)
	case "file-lines":
		seedFileLinesEntries(t, home)
	case "git-dirs":
		if !gitAvailable(t) {
			return
		}
		seedGitDirsEntries(t, home)
	case "node-modules":
		seedNodeModulesEntry(t, home)
	case "entry-order":
		seedEntryOrderEntries(t, home)
	case "on-entry":
		seedOnEntryEntries(t, home)
	default:
		t.Fatalf("unknown SeedProfile %q", profile)
	}
}

func findEntry(t *testing.T, entries []analyse.EntryResult, name string) analyse.EntryResult {
	t.Helper()
	for _, e := range entries {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("entry %q not found in %d results", name, len(entries))
	return analyse.EntryResult{}
}

func assertSortedNames(t *testing.T, names []string) {
	t.Helper()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Fatalf("names not sorted: %v", names)
		}
	}
}

func childNames(entry analyse.EntryResult) []string {
	var names []string
	for _, c := range entry.Children {
		names = append(names, c.Name)
	}
	return names
}

func semanticByKey(entry analyse.EntryResult, key string) (analyse.SemanticLine, bool) {
	for _, line := range entry.Semantic {
		if line.Key == key {
			return line, true
		}
	}
	return analyse.SemanticLine{}, false
}

func Setup(t *testing.T, req *Request) error {
	if req.Mode == "" {
		req.Mode = "scan"
	}
	return nil
}
```