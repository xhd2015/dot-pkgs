## Expected

- `err` is nil.
- Exactly one repo discovered from the valid root.
- Exactly one `RootError` for the missing root path.

## Errors

- Scan returns fatal error instead of partial result.
- Missing repo from good root or missing RootError for bad root.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected nil scan error, got %v", err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(resp.Repos), resp.Repos)
	}
	if len(resp.RootErrors) != 1 {
		t.Fatalf("expected 1 root error, got %d: %v", len(resp.RootErrors), resp.RootErrors)
	}

	wantPath := absPath(t, req.Roots[0])
	r := resp.Repos[0]
	if r.Path != wantPath {
		t.Fatalf("Path = %q, want %q", r.Path, wantPath)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}

	missing := req.Roots[1]
	re := resp.RootErrors[0]
	if re.Root != missing {
		t.Fatalf("RootError.Root = %q, want %q", re.Root, missing)
	}
	if !strings.Contains(re.Error, filepath.Base(missing)) && !strings.Contains(re.Error, missing) {
		t.Fatalf("RootError.Error should mention missing root %q, got %q", missing, re.Error)
	}
}
```