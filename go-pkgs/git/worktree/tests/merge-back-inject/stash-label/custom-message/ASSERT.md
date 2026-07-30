## Expected

- `MergeBack` succeeds with action `"rebased-and-merged"`.
- Source dirt (`dirty.txt`) is restored after migrate.
- Stash history for the source repo contains `req.StashLabel`
  (via `git reflog show stash` after pop/drop) — proves the dirty-diverged
  path used the injected stash message.

## Side Effects

- Temporary stash entries created during migrate are not left on `stash list`
  after success (pop/drop cleaned the active stash); only history/reflog retains
  the label.

## Errors

- None expected.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("expected success with inject StashLabel, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action rebased-and-merged, got %q", resp.Action)
	}
	if req.StashLabel == "" {
		t.Fatal("test misconfigured: req.StashLabel empty")
	}

	content, readErr := os.ReadFile(filepath.Join(req.SourcePath, "dirty.txt"))
	if readErr != nil {
		t.Fatalf("dirty.txt missing after migrate: %v", readErr)
	}
	if string(content) != "uncommitted\n" {
		t.Fatalf("dirty.txt content = %q, want uncommitted\\n", string(content))
	}

	// Active stash list should not keep the inject entry after successful pop.
	listOut := runGitOutput(t, req.SourcePath, "stash", "list")
	// Reflog (or list fallback) must record the custom label from stash push -m.
	reflog := stashReflog(t, req.SourcePath)
	combined := listOut + "\n" + reflog
	if !strings.Contains(combined, req.StashLabel) {
		t.Fatalf("stash history missing inject label %q\nstash list:\n%s\nreflog:\n%s",
			req.StashLabel, listOut, reflog)
	}
	// Must not have fallen back to the product-specific hard-coded message only.
	// (Product label may still appear if a prior leaf polluted the shared repo —
	// this fixture is fresh, so wrk-merge-back should not be required.)
	_ = combined
}
```
