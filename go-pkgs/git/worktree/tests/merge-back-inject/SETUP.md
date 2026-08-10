# Scenario

**Feature**: shared fixture for MergeBack inject-option doctests

```
# every leaf: diverged dirty linked worktree → MergeBack with inject opts
Setup -> main + feature worktree (diverged) + dirty source
  -> Run(MergeBackOptions{TmpDir, StashLabel}) -> Response
```

## Preconditions

- `git` available on PATH.
- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree` importable.
- Classic TDD: `MergeBackOptions.TmpDir` / `StashLabel` may be missing until
  implementer adds them (compile RED OK).
- **No** `os.Setenv` / `t.Setenv` for `WRK_HOME` — inject options replace env.

## Steps

1. Root verifies `git` is on PATH (skip if missing).
2. Root allocates `WorkRoot` via `t.TempDir()`.
3. Leaves build a diverged dirty fixture and set inject fields on `Request`.
4. Root `Run` calls `MergeBack` with `TmpDir` / `StashLabel` from `Request`.

## Context

- Module: `github.com/xhd2015/dot-pkgs/go-pkgs`
- Sibling tree `merge-back-dirty-rebase` covers conflict matrices and still uses
  env-based `WRK_HOME` in older leaves; this tree is focused inject-only and
  must not rewrite that tree.
- Parallel-safe: paths absolute from `t.TempDir()` / `d`; no process cwd/env
  mutation.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	req.WorkRoot = t.TempDir()
	return nil
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, ee.Stderr)
		}
		t.Fatalf("git %v in %s: %v", args, dir, err)
	}
	return strings.TrimSpace(string(out))
}

func initRepo(t *testing.T, path string) {
	t.Helper()
	runGit(t, path, "init", "-b", "master")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(path, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", "init")
}

func addWorktree(t *testing.T, mainRepo, wtDir, branch string) {
	t.Helper()
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
}

// setupDivergedDirty builds main + feature worktree, diverges commits, and
// leaves uncommitted dirt on the feature worktree. Sets MainRepo/SourcePath.
func setupDivergedDirty(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatal(err)
	}
	initRepo(t, mainRepo)

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")
	if err := os.WriteFile(filepath.Join(featureWT, "feature-only.txt"), []byte("f\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, featureWT, "add", "feature-only.txt")
	runGit(t, featureWT, "commit", "-m", "feature only")

	// diverge: commit on main only
	if err := os.WriteFile(filepath.Join(mainRepo, "main-only.txt"), []byte("m\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, mainRepo, "add", "main-only.txt")
	runGit(t, mainRepo, "commit", "-m", "main only")

	// dirty on source (uncommitted)
	if err := os.WriteFile(filepath.Join(featureWT, "dirty.txt"), []byte("uncommitted\n"), 0644); err != nil {
		t.Fatal(err)
	}

	req.MainRepo = mainRepo
	req.SourcePath = featureWT
	req.TargetPath = ""
}

func hasDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isClean(t *testing.T, dir string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status in %s: %v", dir, err)
	}
	return len(strings.TrimSpace(string(out))) == 0
}

func canonPath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	dir, base := filepath.Dir(abs), filepath.Base(abs)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(resolvedDir, base)
	}
	return filepath.Clean(abs)
}

func pathUnder(t *testing.T, child, parent string) bool {
	t.Helper()
	c := canonPath(t, child)
	p := canonPath(t, parent)
	if c == p {
		return true
	}
	prefix := p + string(filepath.Separator)
	return strings.HasPrefix(c, prefix)
}

func stashReflog(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "reflog", "show", "stash")
	out, err := cmd.CombinedOutput()
	// empty stash reflog may exit non-zero on some git versions
	return string(out) + errString(err)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func listDirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
```
