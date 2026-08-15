# Scenario

**Feature**: `Options.Debug` emits greppable `scan:` phase timing and warm/cold mode logs

```
# debug pipeline (library)
caller roots + CacheRoot + Debug
  -> Scan (warm or cold)
  -> when Debug=true: stderrWriter(opts.Stderr) <- scan: lines
     (cacheRoot, mode=warm|cold + reason, serve/refresh/total)
  -> when Debug=false: zero scan: lines

# tests always capture Stderr via bytes.Buffer in Run
```

## Preconditions

- Nested root: does not inherit parent helpers; provides own absPath / fixtures.
- Explicit temp `CacheRoot` only (never `$HOME/.cache/git-repo-scan`).
- `Options.Debug` is a library field; `Run` passes `Debug` and a capture buffer as `Options.Stderr`.
- Fake `.git` fixtures; no enrichment.

## Steps

1. Allocate temp `CacheRoot`.
2. Default `NoCache=false`, `Debug=false` (on/ overrides true).
3. Provide `fakeGitRepo` and `coldSeedScan` for warm branch.

```go
import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

// coldSeedScan populates home/repos.json for warm eligibility (index-only).
// Seed runs without Debug so seed logs never pollute the Scan under test.
func coldSeedScan(t *testing.T, roots []string, cacheRoot string) {
	t.Helper()
	if cacheRoot == "" {
		t.Fatal("coldSeedScan: empty cacheRoot")
	}
	if len(roots) == 0 {
		t.Fatal("coldSeedScan: empty roots")
	}
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     roots,
		CacheRoot: cacheRoot,
		NoCache:   false,
		Debug:     false,
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
	idx, ok, loadErr := scan_repo.LoadRepoIndex(cacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("cold seed LoadRepoIndex: %v", loadErr)
	}
	if !ok || len(idx.Repos) == 0 {
		t.Fatalf("cold seed: expected non-empty home/repos.json under %s", cacheRoot)
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CacheRoot = t.TempDir()
	req.NoCache = false
	req.Debug = false
	return nil
}
```
