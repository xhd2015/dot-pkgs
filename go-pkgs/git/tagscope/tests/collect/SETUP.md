# Scenario

**Feature**: tagscope Phase 1 collects parsed tags and per-scope lineage

```
tag names -> ParseTagName | CollectFromNames | Collect(git tag -l) -> inventory + lineage
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope` is importable.
- Most leaves inject explicit tag name lists (no git subprocess).
- `collect/from-git-repo` requires real `git` on PATH.

## Context

- Pure parse and `CollectFromNames` paths need no filesystem fixtures.
- Git helpers create minimal repos under `t.TempDir()` when needed.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}

func gitAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
		return false
	}
	return true
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func gitInitRepo(t *testing.T, dir string) {
	t.Helper()
	mkdirAll(t, dir)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
}

func gitInitialCommit(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "README"), "init\n")
	runGit(t, dir, "add", "README")
	runGit(t, dir, "commit", "-m", "init")
}

func gitCreateTag(t *testing.T, dir, name string) {
	t.Helper()
	runGit(t, dir, "tag", name)
}

func requireParseOK(t *testing.T, ok bool) {
	t.Helper()
	if !ok {
		t.Fatal("expected parse ok")
	}
}

func requireParseNotOK(t *testing.T, ok bool) {
	t.Helper()
	if ok {
		t.Fatal("expected parse not ok")
	}
}

func scopeKey(scope tagscope.TagScope) tagscope.TagScopeKey {
	return tagscope.TagScopeKey(scope.PathPrefix)
}

func lineageFor(t *testing.T, collected tagscope.CollectedTags, prefix string) tagscope.ScopeLineage {
	t.Helper()
	key := tagscope.TagScopeKey(prefix)
	lineage, ok := collected.ByScope[key]
	if !ok {
		t.Fatalf("scope %q missing from ByScope", prefix)
	}
	return lineage
}
```