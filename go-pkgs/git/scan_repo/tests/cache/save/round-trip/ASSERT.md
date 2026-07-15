## Expected

- `err` is nil; `resp.EntryOK` is true.
- Loaded `resp.Entry` matches saved fields:
  `Version`, `RefreshedAt`, `MtimeNs`, `IsRepo`, `RepoType`, `GitDir`,
  `Children`, `ScanComplete`, `OptionsHash`.
- On disk, `entry.json` exists at the expected mirror path (nested dirs created).

## Errors

- `err` is nil.

## Side Effects

- File exists at
  `<CacheRoot>/mirror/Users/xhd2015/Projects/org/saved-repo/entry.json`.

```go
import (
	"os"
	"path/filepath"
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
	if !resp.EntryOK {
		t.Fatal("expected EntryOK true after save-load")
	}
	got, want := resp.Entry, req.Entry
	if got.Version != want.Version {
		t.Fatalf("Version = %d, want %d", got.Version, want.Version)
	}
	if got.RefreshedAt != want.RefreshedAt {
		t.Fatalf("RefreshedAt = %q, want %q", got.RefreshedAt, want.RefreshedAt)
	}
	if got.MtimeNs != want.MtimeNs {
		t.Fatalf("MtimeNs = %d, want %d", got.MtimeNs, want.MtimeNs)
	}
	if got.IsRepo != want.IsRepo {
		t.Fatalf("IsRepo = %v, want %v", got.IsRepo, want.IsRepo)
	}
	if got.RepoType != want.RepoType {
		t.Fatalf("RepoType = %q, want %q", got.RepoType, want.RepoType)
	}
	if got.GitDir != want.GitDir {
		t.Fatalf("GitDir = %q, want %q", got.GitDir, want.GitDir)
	}
	if !reflect.DeepEqual(got.Children, want.Children) {
		t.Fatalf("Children = %v, want %v", got.Children, want.Children)
	}
	if got.ScanComplete != want.ScanComplete {
		t.Fatalf("ScanComplete = %v, want %v", got.ScanComplete, want.ScanComplete)
	}
	if got.OptionsHash != want.OptionsHash {
		t.Fatalf("OptionsHash = %q, want %q", got.OptionsHash, want.OptionsHash)
	}

	path := expectedMirrorEntryPath(t, req.CacheRoot, req.RealPath)
	st, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("expected entry file at %s: %v", path, statErr)
	}
	if st.IsDir() {
		t.Fatalf("entry path is a directory: %s", path)
	}
	// Nested parents under mirror must exist (multi-segment real path).
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("expected nested mirror dir: %v", err)
	}
}
```
