## Expected

- `MergeBack` succeeds with action `"rebased-and-merged"`.
- `resp.ObservedTmpPath` is non-empty and is a path **under** `req.TmpDir`
  (inject parent used for tmp worktree creation).
- Source worktree still exists and remains dirty (`dirty.txt` present).
- After success, `req.TmpDir` has no leftover tmp worktree entries.

## Side Effects

- Feature branch rebased/merged into target; tmp worktree and tmp branch removed.

## Errors

- None expected.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("expected success with inject TmpDir, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action rebased-and-merged, got %q", resp.Action)
	}
	if resp.ObservedTmpPath == "" {
		t.Fatal("expected ObservedTmpPath from Confirm plan.Commands[0].Dir")
	}
	if req.TmpDir == "" {
		t.Fatal("test misconfigured: req.TmpDir empty")
	}
	if !pathUnder(t, resp.ObservedTmpPath, req.TmpDir) {
		t.Fatalf("ObservedTmpPath %q is not under inject TmpDir %q",
			resp.ObservedTmpPath, req.TmpDir)
	}

	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
	if isClean(t, req.SourcePath) {
		t.Fatal("source should still be dirty after merge-back")
	}
	content, readErr := os.ReadFile(filepath.Join(req.SourcePath, "dirty.txt"))
	if readErr != nil {
		t.Fatalf("dirty.txt missing: %v", readErr)
	}
	if string(content) != "uncommitted\n" {
		t.Fatalf("dirty.txt content = %q, want uncommitted\\n", string(content))
	}

	left := listDirNames(t, req.TmpDir)
	if len(left) > 0 {
		t.Fatalf("tmp worktrees left under inject TmpDir: %v", left)
	}
}
```
