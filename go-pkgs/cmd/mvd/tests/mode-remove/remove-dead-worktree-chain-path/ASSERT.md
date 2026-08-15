## Expected
- Exit code 0; output contains `removed:`.
- wt2 is removed from the chain; root, wt1, and dst remain.
- Root entry is NOT deleted (no `-f` needed).

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "removed:")

	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	assertHistoryChain(t, req.ConfigHome, root, root, wt1, dst)
	assertHistoryWorktreeEntry(t, req.ConfigHome, root, 1, root, "feature-a")

	h := readHistoryFile(t, req.ConfigHome)
	proj := h.Projects[root]
	for _, loc := range proj.Locations {
		if loc.Path == wt2 {
			t.Fatalf("wt2 %s should be removed from chain, got %#v", wt2, proj.Locations)
		}
	}
}
```