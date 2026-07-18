## Expected

- `err` is nil; `resp.IndexOK` is true.
- Loaded `resp.Index` matches saved multi-entry document for universe `root`.
- On disk, `repos.json` exists at `<CacheRoot>/root/repos.json` only (not home).

## Errors

- `err` is nil.

## Side Effects

- File at `<CacheRoot>/root/repos.json`; no accidental write to `home/repos.json`.

```go
import (
	"os"
	"reflect"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	if got.Universe != "root" {
		t.Fatalf("Universe = %q, want root", got.Universe)
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
	if len(got.Repos) != 2 {
		t.Fatalf("Repos len = %d, want 2", len(got.Repos))
	}

	wantPath := expectedRepoIndexPath(t, req.CacheRoot, "root")
	if resp.IndexPath != wantPath {
		t.Fatalf("IndexPath = %q, want %q", resp.IndexPath, wantPath)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected repos.json at %s: %v", wantPath, err)
	}
	homePath := expectedRepoIndexPath(t, req.CacheRoot, "home")
	if _, err := os.Stat(homePath); err == nil {
		t.Fatalf("home/repos.json must not exist after root-only save: %s", homePath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat home path: %v", err)
	}
}
```
