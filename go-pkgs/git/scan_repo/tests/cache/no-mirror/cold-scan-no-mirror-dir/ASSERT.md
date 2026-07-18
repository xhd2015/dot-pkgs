## Expected

- Scan succeeds; discovers `my-repo` as main.
- Path `<CacheRoot>/mirror` does **not** exist (neither as a directory nor a file).
- Optional: `home/repos.json` may be present (index seed is allowed).

## Errors

- `err` is nil.

## Side Effects

- Dense mirror cache is dead; cold Scan must not recreate it under CacheRoot.
- This leaf is **RED** until product stops writing mirror entries on cold Scan.

```go
import (
	"os"
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
	repoPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	if len(resp.Repos) != 1 || resp.Repos[0].Path != repoPath {
		t.Fatalf("expected discovery of %s, got %v", repoPath, resp.Repos)
	}
	if resp.Repos[0].RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", resp.Repos[0].RepoType)
	}

	mirrorDir := filepath.Join(req.CacheRoot, "mirror")
	if st, statErr := os.Stat(mirrorDir); statErr == nil {
		t.Fatalf("cold Scan created mirror path %s (mode=%v); mirror cache is retired — want path absent", mirrorDir, st.Mode())
	} else if !os.IsNotExist(statErr) {
		t.Fatalf("stat mirror: %v", statErr)
	}
}
```
