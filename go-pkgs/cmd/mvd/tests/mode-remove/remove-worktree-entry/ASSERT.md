## Expected
- `--rm wt1` succeeds, removing only the wt1 worktree entry from the chain.
- The output contains "removed:".
- The history still has the repo entry with 2 locations: root and wt2 (wt1 is gone).

## Exit Code
- 0 (success)

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		t.Fatalf("expected response, got error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code: %d, output:\n%s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "removed:")

	repo := filepath.Join(req.WorkRoot, "repo")
	wt2 := filepath.Join(req.WorkRoot, "wt2")

	h := assertHistoryLen(t, req.ConfigHome, 1)
	proj := h.Projects[repo]
	if len(proj.Locations) != 2 {
		t.Fatalf("expected 2 locations (root + wt2), got %d: %#v", len(proj.Locations), proj.Locations)
	}
	if proj.Locations[0].Path != repo {
		t.Fatalf("expected root %s, got %s", repo, proj.Locations[0].Path)
	}
	if proj.Locations[1].Path != wt2 {
		t.Fatalf("expected wt2 %s, got %s", wt2, proj.Locations[1].Path)
	}
}
```
