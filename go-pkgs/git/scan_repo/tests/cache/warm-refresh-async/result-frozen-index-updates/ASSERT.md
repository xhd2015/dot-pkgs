## Expected

- `ScanSession` with `WarmRefreshAsync` returns serve-only Result (known-repo only).
- `Refresh` handle is non-nil.
- After `Join`, durable index includes `new-repo`.
- Result is unchanged after Join (still one repo).

```go
import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	// Isolated fixture: parent Run uses classic Scan and can poison a shared
	// CacheRoot (sync refresh / walk-log). Rebuild for a true async serve snapshot.
	_ = req
	_ = resp
	_ = err

	root := t.TempDir()
	cacheRoot := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	known := filepath.Join(unitA, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)
	coldSeedScan(t, []string{root}, cacheRoot)

	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))

	// Disable walk-log sync discover during serve: stamp last_scan_end to "now"
	// so delta < 10s → 0 walk-log budget; unit refresh still finds nested new.
	if err := scan_repo.SaveLastScanEnd(cacheRoot, time.Now()); err != nil {
		t.Fatalf("SaveLastScanEnd: %v", err)
	}

	sess, sessErr := scan_repo.ScanSession(context.Background(), scan_repo.Options{
		Roots:             []string{root},
		CacheRoot:         cacheRoot,
		NoCache:           false,
		WarmRefreshMode:   scan_repo.WarmRefreshAsync,
		WarmRefreshBudget: 2 * time.Second,
		YoungAge:          0,
		LastScanEnd:       time.Now(),
	})
	if sessErr != nil {
		t.Fatal(sessErr)
	}
	if sess.Refresh == nil {
		t.Fatal("expected non-nil Refresh handle for warm async")
	}

	knownPath := absPath(t, known)
	newPath := absPath(t, newRepo)

	if len(sess.Result.Repos) != 1 {
		t.Fatalf("async Result must be serve-frozen (1 known), got %d: %v",
			len(sess.Result.Repos), pathsOf(sess.Result.Repos))
	}
	if sess.Result.Repos[0].Path != knownPath {
		t.Fatalf("Result[0]=%q want known %q", sess.Result.Repos[0].Path, knownPath)
	}

	joinCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if joinErr := sess.Join(joinCtx); joinErr != nil {
		t.Fatalf("Join: %v", joinErr)
	}

	if len(sess.Result.Repos) != 1 || sess.Result.Repos[0].Path != knownPath {
		t.Fatalf("Result mutated after Join: %v", pathsOf(sess.Result.Repos))
	}

	idx, ok, loadErr := scan_repo.LoadRepoIndex(cacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatal("expected home/repos.json after Join")
	}
	foundNew := false
	for _, e := range idx.Repos {
		if e.Path == newPath {
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Fatalf("index missing new-repo after Join: %s", newPath)
	}
}

func pathsOf(repos []scan_repo.Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Path
	}
	return out
}
```
