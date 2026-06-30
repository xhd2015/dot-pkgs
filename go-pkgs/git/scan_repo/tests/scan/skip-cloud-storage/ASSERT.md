## Expected

- `err` is nil.
- Exactly one repo: `visible-repo`.
- CloudStorage repo is not discovered.

## Errors

- No error returned from `Run`.

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
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(resp.Repos), resp.Repos)
	}
	wantPath := absPath(t, filepath.Join(req.Roots[0], "visible-repo"))
	if resp.Repos[0].Path != wantPath {
		t.Fatalf("Path = %q, want %q", resp.Repos[0].Path, wantPath)
	}
	if resp.Repos[0].RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", resp.Repos[0].RepoType)
	}
}
```