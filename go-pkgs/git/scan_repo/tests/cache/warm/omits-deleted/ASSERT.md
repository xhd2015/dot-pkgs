## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `still-here` (main), correct Path/GitDir.
- `gone-repo` is **not** listed in `resp.Repos`.
- Durable index: gone path absent after warm (liveness drop), still-here present.

## Errors

- `err` is nil.

## Side Effects

- Stale index rows for deleted paths must not survive warm Scan serve.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	stillPath := absPath(t, filepath.Join(req.Roots[0], "still-here"))
	wantGitDir := absPath(t, filepath.Join(stillPath, ".git"))
	gonePath := req.RealPath
	if gonePath == "" {
		t.Fatal("req.RealPath empty; Setup must stash deleted repo abs path")
	}

	for i, r := range resp.Repos {
		if r.Path == gonePath {
			t.Fatalf("warm Scan listed deleted gone-repo at repos[%d]=%q", i, r.Path)
		}
	}

	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 live repo (still-here), got %d", len(resp.Repos))
	}
	r := resp.Repos[0]
	if r.Path != stillPath {
		t.Fatalf("Repos[0].Path = %q, want %q", r.Path, stillPath)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	if r.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGitDir)
	}

	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatal("expected home/repos.json after warm Scan")
	}
	for _, e := range idx.Repos {
		if e.Path == gonePath {
			t.Fatalf("index still lists deleted path %s after warm liveness", gonePath)
		}
	}
	foundStill := false
	for _, e := range idx.Repos {
		if e.Path == stillPath {
			foundStill = true
			break
		}
	}
	if !foundStill {
		t.Fatalf("index missing still-here %s", stillPath)
	}
}
```
