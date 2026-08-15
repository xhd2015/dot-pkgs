---
label: unix
explanation: chmod 000 directory fixture; library Scan walk robustness
---

## Expected

- `err` is nil — walk does not fail on permission denied under a sibling directory.
- Exactly one repo: `visible-repo`.

## Errors

- No error returned from `Run`.

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
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(resp.Repos), resp.Repos)
	}
	wantPath := absPath(t, filepath.Join(req.Roots[0], "visible-repo"))
	r := resp.Repos[0]
	if r.Path != wantPath {
		t.Fatalf("Path = %q, want %q", r.Path, wantPath)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
}
```