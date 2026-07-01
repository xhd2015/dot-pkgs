# Scenario

**Feature**: scan walks a root dir, applies skip rules, and returns/streams Go modules

```
# caller hands scan a root workspace (go.mod files + optional git)
root + nested go.mod files (+ .gitignore / nested .git) -> scan walker -> Modules

# skip rules prune .git/vendor/testdata, gitignored dirs, nested separate repos
walker -> name skips + git skips -> go.mod reader -> Module{Dir,Path,Requires,Replaces}

# two consumption modes: Scan (sorted batch) vs ScanStream (walk order, unsorted)
caller -> Scan(root)        -> sorted []Module
caller -> ScanStream(root)  -> per-module fn in walk order
```

## Preconditions

- The `scan` package is importable at `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan`
  (not yet implemented — leaves are RED until it lands).
- `go` and `git` are available on PATH (git only needed for leaves that init a repo).

## Steps

1. Verify `go` and `git` are available.
2. Leaf `Setup` builds an isolated temp workspace: `go.mod` files + git init/config/add/commit
   and/or `.gitignore` as the scenario requires.
3. Leaf `Setup` sets `req.Operation` (`"scan"` or `"stream"`) and `req.RootDir`.
4. Root `Run` dispatches to `scan.Scan` or `scan.ScanStream`.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Setup verifies tooling. Leaf Setup functions build fixtures and set req fields.
func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	return nil
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// newWorkspace returns an isolated temp dir, cleaned up after the test.
func newWorkspace(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "scan-doctest-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// writeModule writes a go.mod declaring modulePath with go 1.22. If requires is
// non-empty each "path=version" entry is added as a require; this lets leaves
// exercise the Requires/Replaces fields without running `go mod` (keeps fixtures
// hermetic and offline).
func writeModule(t *testing.T, dir, modulePath string, requires ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("module " + modulePath + "\n\ngo 1.22\n")
	if len(requires) > 0 {
		b.WriteString("\nrequire (\n")
		for _, r := range requires {
			b.WriteString("\t" + r + "\n")
		}
		b.WriteString(")\n")
	}
	writeFile(t, filepath.Join(dir, "go.mod"), b.String())
}

// initGitRepo runs git init + identity config + add + commit in dir, so dir becomes
// the root git repo (used for the workspace root or for a nested separate repo).
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test User")
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "init")
}

// dirLines returns the Dir fields of ms as a slice, for assertions.
func dirLines(ms []scan.Module) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.Dir)
	}
	return out
}

// pathOf returns the module Path for the module whose Dir equals dir, or "" if absent.
func pathOf(ms []scan.Module, dir string) string {
	for _, m := range ms {
		if m.Dir == dir {
			return m.Path
		}
	}
	return ""
}

// hasDir reports whether any module has Dir == dir.
func hasDir(ms []scan.Module, dir string) bool {
	for _, m := range ms {
		if m.Dir == dir {
			return true
		}
	}
	return false
}
```
