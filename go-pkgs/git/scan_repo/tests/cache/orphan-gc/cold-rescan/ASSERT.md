## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `keep` (main), correct Path/GitDir.
- `gone` is **not** listed in `resp.Repos`.
- Mirror for deleted path (`req.RealPath`): **Load returns ok=false**
  (orphan GC removed the entry/subtree — stronger than warm liveness
  leaving `is_repo=false`).
- Scan-root mirror `children` includes `"keep"` and does **not** include `"gone"`.
- Mirror for `keep` still has `is_repo=true`.

## Errors

- `err` is nil.

## Side Effects

- After cold force rewalk of the parent, dead child mirror path is gone so the
  cache cannot accumulate forever when directories are deleted.

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
	if !req.Refresh {
		t.Fatal("req.Refresh must be true for cold-rescan orphan GC")
	}

	keepPath := absPath(t, filepath.Join(req.Roots[0], "keep"))
	wantGitDir := absPath(t, filepath.Join(keepPath, ".git"))
	gonePath := req.RealPath
	if gonePath == "" {
		t.Fatal("req.RealPath empty; Setup must stash deleted gone abs path")
	}

	for i, r := range resp.Repos {
		if r.Path == gonePath {
			t.Fatalf("cold rescan listed deleted gone at repos[%d]=%q", i, r.Path)
		}
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 live repo (keep), got %d", len(resp.Repos))
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

	// P7 orphan GC: entry must be removed (not merely is_repo=false).
	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, gonePath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(gone): %v", loadErr)
	}
	if ok {
		t.Fatalf("orphan mirror still present for gone (want Load ok=false after GC); entry=%+v", entry)
	}

	// Surviving sibling still cached.
	keepEntry, keepOK, keepErr := scan_repo.LoadCacheEntry(req.CacheRoot, keepPath)
	if keepErr != nil {
		t.Fatalf("LoadCacheEntry(keep): %v", keepErr)
	}
	if !keepOK {
		t.Fatal("expected mirror entry for keep after cold rescan")
	}
	if !keepEntry.IsRepo {
		t.Fatalf("keep IsRepo=false, want true; entry=%+v", keepEntry)
	}

	// Parent children rewritten without the orphan basename.
	rootPath := absPath(t, req.Roots[0])
	rootEntry, rootOK, rootErr := scan_repo.LoadCacheEntry(req.CacheRoot, rootPath)
	if rootErr != nil {
		t.Fatalf("LoadCacheEntry(root): %v", rootErr)
	}
	if !rootOK {
		t.Fatal("expected mirror entry for scan root after cold rescan")
	}
	hasKeep, hasGone := false, false
	for _, c := range rootEntry.Children {
		if c == "keep" {
			hasKeep = true
		}
		if c == "gone" {
			hasGone = true
		}
	}
	if !hasKeep {
		t.Fatalf("root Children = %v, want to include %q", rootEntry.Children, "keep")
	}
	if hasGone {
		t.Fatalf("root Children = %v, must not include orphan %q", rootEntry.Children, "gone")
	}
}
```
