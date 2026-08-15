## Expected

- `err` is nil; `resp.IndexOK` is true.
- Loaded `resp.Index` matches saved fields:
  `Version`, `Universe`, `Base`, `UpdatedAt`, and each `Repos[]` entry's
  `Path`, `RepoType`, `GitDir`, `Depth`, `SeenAt`.
- On disk, `repos.json` exists at `<CacheRoot>/home/repos.json`.

## Errors

- `err` is nil.

## Side Effects

- File exists at `resp.IndexPath` (and equals expected home path).

```go
import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if !resp.IndexOK {
		t.Fatal("expected IndexOK true after save-load")
	}
	want := req.Index
	got := resp.Index
	if got.Version != want.Version {
		t.Fatalf("Version = %d, want %d", got.Version, want.Version)
	}
	if got.Universe != want.Universe {
		t.Fatalf("Universe = %q, want %q", got.Universe, want.Universe)
	}
	if got.Base != want.Base {
		t.Fatalf("Base = %q, want %q", got.Base, want.Base)
	}
	if got.UpdatedAt != want.UpdatedAt {
		t.Fatalf("UpdatedAt = %q, want %q", got.UpdatedAt, want.UpdatedAt)
	}
	if !reflect.DeepEqual(got.Repos, want.Repos) {
		t.Fatalf("Repos = %+v, want %+v", got.Repos, want.Repos)
	}

	wantPath := expectedRepoIndexPath(t, req.CacheRoot, "home")
	if resp.IndexPath != wantPath {
		t.Fatalf("IndexPath = %q, want %q", resp.IndexPath, wantPath)
	}
	st, statErr := os.Stat(wantPath)
	if statErr != nil {
		t.Fatalf("expected repos.json at %s: %v", wantPath, statErr)
	}
	if st.IsDir() {
		t.Fatalf("repos.json path is a directory: %s", wantPath)
	}
	if _, err := os.Stat(filepath.Dir(wantPath)); err != nil {
		t.Fatalf("expected universe dir: %v", err)
	}
}
```
