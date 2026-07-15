## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `unit-a/keep` (main).
- Deleted `gone` is **not** listed.
- Mirror for deleted path: **Load returns ok=false** (orphan GC under unit rewalk).
- Unit parent `unit-a` mirror `children` includes `"keep"` and does **not** include `"gone"`.
- Mirror for `unit-a/keep` still has `is_repo=true`.

## Errors

- `err` is nil.

## Side Effects

- Budgeted warm unit rewalk rewrites the unit parent and drops dead child mirrors
  (same GC contract as cold rescan, scoped to the rewalked unit).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}
	if req.Refresh {
		t.Fatal("req.Refresh must be false for unit-refresh orphan GC")
	}

	keepPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "keep"))
	wantGitDir := absPath(t, filepath.Join(keepPath, ".git"))
	gonePath := req.RealPath
	if gonePath == "" {
		t.Fatal("req.RealPath empty; Setup must stash deleted gone abs path")
	}

	for i, r := range resp.Repos {
		if r.Path == gonePath {
			t.Fatalf("unit refresh listed deleted gone at repos[%d]=%q", i, r.Path)
		}
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 live repo (unit-a/keep), got %d", len(resp.Repos))
	}
	r := resp.Repos[0]
	if r.Path != keepPath {
		t.Fatalf("Repos[0].Path = %q, want %q", r.Path, keepPath)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	if r.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGitDir)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, gonePath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(gone): %v", loadErr)
	}
	if ok {
		t.Fatalf("orphan mirror still present for gone (want Load ok=false after unit GC); entry=%+v", entry)
	}

	keepEntry, keepOK, keepErr := scan_repo.LoadCacheEntry(req.CacheRoot, keepPath)
	if keepErr != nil {
		t.Fatalf("LoadCacheEntry(keep): %v", keepErr)
	}
	if !keepOK {
		t.Fatal("expected mirror entry for keep after unit refresh")
	}
	if !keepEntry.IsRepo {
		t.Fatalf("keep IsRepo=false, want true; entry=%+v", keepEntry)
	}

	unitPath := absPath(t, filepath.Join(req.Roots[0], "unit-a"))
	unitEntry, unitOK, unitErr := scan_repo.LoadCacheEntry(req.CacheRoot, unitPath)
	if unitErr != nil {
		t.Fatalf("LoadCacheEntry(unit-a): %v", unitErr)
	}
	if !unitOK {
		t.Fatal("expected mirror entry for unit-a after unit refresh")
	}
	hasKeep, hasGone := false, false
	for _, c := range unitEntry.Children {
		if c == "keep" {
			hasKeep = true
		}
		if c == "gone" {
			hasGone = true
		}
	}
	if !hasKeep {
		t.Fatalf("unit-a Children = %v, want to include %q", unitEntry.Children, "keep")
	}
	if hasGone {
		t.Fatalf("unit-a Children = %v, must not include orphan %q", unitEntry.Children, "gone")
	}
}
```
