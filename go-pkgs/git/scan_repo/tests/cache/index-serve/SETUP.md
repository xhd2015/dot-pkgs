# Scenario

**Feature**: Scan seeds and serves per-universe `repos.json` (P2 index serve + sibling)

```
# cold: discover repos and SaveRepoIndex(universe=home, Base=absRoot)
caller roots + CacheRoot + NoCache=false
  -> Scan cold walk
  -> <CacheRoot>/home/repos.json lists main paths

# warm: serve from index + liveness + sibling ReadDir
prior cold left home/repos.json
  -> Scan warm
  -> Result from index candidates (live .git)
  -> sibling of indexed repo at parent/B with .git is discovered without Refresh
  -> dead indexed path (no .git) omitted
```

## Preconditions

- Nested root: own helpers; does not inherit parent tree `Setup`.
- Every leaf uses an explicit temp `CacheRoot` from `t.TempDir()` (never
  `$HOME/.cache/git-repo-scan`).
- P2 simplest universe: always load/assert `"home"` after Scan; `Base` is abs of
  the single scan root.
- Fake `.git` fixtures only; `ListRemotes` / `ListWorktrees` off.
- Cold seed helper asserts warm eligibility via `LoadRepoIndex(home)` (entries
  under the scan root); dense mirror is retired.

## Steps

1. Allocate a fresh temp `CacheRoot`.
2. Default `NoCache=false`, `Refresh=false`, `Debug=false`.
3. Provide `fakeGitRepo`, `absPath`, `mkdirAll`, `coldSeedScan` for descendants.

```go
import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

// fakeGitRepo plants a minimal main-repo .git directory (objects only).
func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

// coldSeedScan runs a full cold Scan that seeds home/repos.json under CacheRoot.
// Leaves may mutate the workspace afterward; Run exercises the warm/index path.
// Warm eligibility is index-only (no mirror).
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
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
	// Sanity: home index must exist for warm (index-only warm eligibility).
	rootPath := absPath(t, roots[0])
	idx, ok, loadErr := scan_repo.LoadRepoIndex(cacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("cold seed LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatalf("cold seed: expected home/repos.json under %s", cacheRoot)
	}
	nUnder := 0
	for _, e := range idx.Repos {
		if e.Path == rootPath || strings.HasPrefix(e.Path, rootPath+string(filepath.Separator)) {
			nUnder++
		}
	}
	if nUnder == 0 {
		t.Fatalf("cold seed: index has no entries under scan root %s; idx=%+v", rootPath, idx)
	}
}

// indexPaths returns the set of RepoIndexEntry.Path values.
func indexPaths(idx scan_repo.RepoIndex) map[string]struct{} {
	out := make(map[string]struct{}, len(idx.Repos))
	for _, e := range idx.Repos {
		out[e.Path] = struct{}{}
	}
	return out
}

// resultPaths returns the set of Result repo Paths.
func resultPaths(repos []scan_repo.Repo) map[string]struct{} {
	out := make(map[string]struct{}, len(repos))
	for _, r := range repos {
		out[r.Path] = struct{}{}
	}
	return out
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CacheRoot = t.TempDir()
	req.NoCache = false
	req.Refresh = false
	req.Debug = false
	return nil
}
```
