## Expected

- Scan succeeds with no RootErrors.
- Exactly two repos: `A` and `B` (path-sorted: A then B), both `RepoTypeMain`.
- `req.Refresh` is false — discovery must not rely on force-cold.
- Proves sibling probe: uncached `B` next to indexed `A` is found on warm Scan.
- RED while warm only serves mirror `is_repo` marks and omits brand-new.

## Errors

- `err` is nil.

## Side Effects

- Optional: after warm, index may gain `B` on save-back; not required for this
  leaf — discovery in `Result.Repos` is the contract.

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
		t.Fatal("Refresh must be false; sibling discovery is not force-cold")
	}

	a := req.KnownPath
	b := req.SiblingPath
	if a == "" || b == "" {
		t.Fatal("Setup must set KnownPath (A) and SiblingPath (B)")
	}

	got := resultPaths(resp.Repos)
	if _, ok := got[a]; !ok {
		t.Fatalf("Result missing cold-indexed A %q; repos=%v", a, resp.Repos)
	}
	if _, ok := got[b]; !ok {
		t.Fatalf("Result missing sibling B %q (want sibling ReadDir discovery); repos=%v", b, resp.Repos)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected exactly 2 repos (A+B), got %d", len(resp.Repos))
	}
	// Path-sorted
	if resp.Repos[0].Path != a || resp.Repos[1].Path != b {
		t.Fatalf("paths = [%q, %q], want [%q, %q]",
			resp.Repos[0].Path, resp.Repos[1].Path, a, b)
	}
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
		wantGit := absPath(t, filepath.Join(r.Path, ".git"))
		if r.GitDir != wantGit {
			t.Fatalf("repos[%d].GitDir = %q, want %q", i, r.GitDir, wantGit)
		}
	}
}
```
