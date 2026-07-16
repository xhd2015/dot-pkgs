## Expected

- Scan succeeds with no RootErrors; `my-repo` discovered.
- `resp.Stderr` contains **zero** occurrences of the substring `scan:`
  (case-sensitive greppable prefix for debug phase logs).
- Existing Verbose skip-warning format is out of scope here (`Verbose` false);
  this leaf only guards against Debug log spam when Debug is off.

## Errors

- `err` is nil.

## Side Effects

- Cold cache write may still occur; discovery and silence of `scan:` are the asserts.

```go
import (
	"path/filepath"
	"strings"
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
	if req.Debug {
		t.Fatal("req.Debug must be false for off leaf")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	repoPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	var found bool
	for _, r := range resp.Repos {
		if r.Path == repoPath && r.RepoType == scan_repo.RepoTypeMain {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected discovery of main repo %s, got %v", repoPath, resp.Repos)
	}

	if strings.Contains(resp.Stderr, "scan:") {
		t.Fatalf("Debug=false must emit zero scan: markers, got stderr:\n%s", resp.Stderr)
	}
}
```
