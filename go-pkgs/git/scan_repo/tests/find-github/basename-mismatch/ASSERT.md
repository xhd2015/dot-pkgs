## Expected

- `resp.Found.Path` is the `myproject-clone` main checkout (absolute).
- `resp.Found.RepoType` is main.
- `resp.Found.Name` is `myproject-clone` (local basename; callers use github name separately).

## Errors

- `err` is nil.

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
	if resp.Found == nil {
		t.Fatal("expected found repo")
	}
	wantPath := absPath(t, filepath.Join(req.Roots[0], "myproject-clone"))
	if resp.Found.Path != wantPath {
		t.Fatalf("Path = %q, want %q", resp.Found.Path, wantPath)
	}
	if resp.Found.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", resp.Found.RepoType)
	}
	if resp.Found.Name != "myproject-clone" {
		t.Fatalf("Name = %q, want myproject-clone", resp.Found.Name)
	}
}
```