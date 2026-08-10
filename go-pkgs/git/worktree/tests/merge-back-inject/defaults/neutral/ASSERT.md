## Expected

- `MergeBack` succeeds with action `"rebased-and-merged"` when inject fields are
  empty (library defaults).
- `resp.ObservedTmpPath` is non-empty and is under `os.TempDir()` (neutral
  default parent), **not** under `$HOME/.wrk`.
- Source dirt restored.
- Stash history for this run does **not** use the product-specific message
  `"wrk-merge-back"` (default stash label is product-neutral).

## Side Effects

- Tmp worktree cleaned up from the neutral parent.

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
	if err != nil {
		t.Fatalf("expected success with neutral defaults, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action rebased-and-merged, got %q", resp.Action)
	}
	if req.TmpDir != "" || req.StashLabel != "" {
		t.Fatalf("defaults leaf must use empty inject fields; got TmpDir=%q StashLabel=%q",
			req.TmpDir, req.StashLabel)
	}
	if resp.ObservedTmpPath == "" {
		t.Fatal("expected ObservedTmpPath from Confirm plan (default tmp location)")
	}

	// Neutral tmp base: under os.TempDir(), not $HOME/.wrk
	if !pathUnder(t, resp.ObservedTmpPath, os.TempDir()) {
		t.Fatalf("default ObservedTmpPath %q is not under os.TempDir() %q (not neutral)",
			resp.ObservedTmpPath, os.TempDir())
	}
	if home, homeErr := os.UserHomeDir(); homeErr == nil {
		wrkWorktrees := filepath.Join(home, ".wrk", "worktrees")
		if pathUnder(t, resp.ObservedTmpPath, wrkWorktrees) ||
			pathUnder(t, resp.ObservedTmpPath, filepath.Join(home, ".wrk")) {
			t.Fatalf("default ObservedTmpPath %q must not live under ~/.wrk (product-specific)",
				resp.ObservedTmpPath)
		}
	}

	content, readErr := os.ReadFile(filepath.Join(req.SourcePath, "dirty.txt"))
	if readErr != nil {
		t.Fatalf("dirty.txt missing: %v", readErr)
	}
	if string(content) != "uncommitted\n" {
		t.Fatalf("dirty.txt content = %q", string(content))
	}

	listOut := runGitOutput(t, req.SourcePath, "stash", "list")
	reflog := stashReflog(t, req.SourcePath)
	combined := listOut + "\n" + reflog
	if strings.Contains(combined, "wrk-merge-back") {
		t.Fatalf("neutral default stash label must not be product-specific %q; stash history:\n%s",
			"wrk-merge-back", combined)
	}
}
```
