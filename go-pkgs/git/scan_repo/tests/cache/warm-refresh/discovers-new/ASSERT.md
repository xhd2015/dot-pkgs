## Expected

- Scan succeeds with no RootErrors.
- Exactly two repos, path-sorted: `known-repo` then `new-repo` under `unit-a`.
- Both are `RepoTypeMain` with correct Path/GitDir.
- Index has an entry for `new-repo` after budgeted refresh.
- No `mirror/` under CacheRoot.

## Errors

- `err` is nil.

## Side Effects

- Refresh rewalk of aged unit merges new repo into Result and home/repos.json.

```go
import (
	"os"
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

	knownPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "known-repo"))
	newPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "nested", "new-repo"))
	wantKnownGit := absPath(t, filepath.Join(knownPath, ".git"))
	wantNewGit := absPath(t, filepath.Join(newPath, ".git"))

	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos (known + refreshed new), got %d: %v", len(resp.Repos), pathsOf(resp.Repos))
	}
	if resp.Repos[0].Path != knownPath {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, knownPath)
	}
	if resp.Repos[1].Path != newPath {
		t.Fatalf("repos[1].Path = %q, want %q (budgeted refresh must discover new under aged unit)", resp.Repos[1].Path, newPath)
	}
	if resp.Repos[0].RepoType != scan_repo.RepoTypeMain || resp.Repos[0].GitDir != wantKnownGit {
		t.Fatalf("known-repo shape: type=%v gitDir=%q", resp.Repos[0].RepoType, resp.Repos[0].GitDir)
	}
	if resp.Repos[1].RepoType != scan_repo.RepoTypeMain || resp.Repos[1].GitDir != wantNewGit {
		t.Fatalf("new-repo shape: type=%v gitDir=%q", resp.Repos[1].RepoType, resp.Repos[1].GitDir)
	}

	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatal("expected home/repos.json after refresh")
	}
	foundNew := false
	for _, e := range idx.Repos {
		if e.Path == newPath {
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Fatalf("index missing new-repo after refresh: %s", newPath)
	}

	mirrorDir := filepath.Join(req.CacheRoot, "mirror")
	if _, err := os.Stat(mirrorDir); err == nil {
		t.Fatalf("mirror path exists at %s; dense mirror is retired", mirrorDir)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat mirror: %v", err)
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
