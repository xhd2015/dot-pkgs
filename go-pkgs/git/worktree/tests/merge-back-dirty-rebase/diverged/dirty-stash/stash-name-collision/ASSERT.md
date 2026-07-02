## Expected

- `MergeBack` succeeds, action `"rebased-and-merged"`.
- `fresh-dirty.txt` survives in source (new dirty change, not the old stashed one).
- The old stash is NOT consumed — it still exists after merge-back.
- No cross-contamination from the old stash.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success with stash name collision, got: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	// fresh-dirty.txt must survive
	content, _ := os.ReadFile(filepath.Join(req.SourcePath, "fresh-dirty.txt"))
	if string(content) != "fresh content\n" {
		t.Fatalf("fresh-dirty.txt content wrong: %q", string(content))
	}

	// Old stash "wrk-merge-back" should still exist (was not consumed)
	cmd := exec.Command("git", "-C", req.SourcePath, "stash", "list")
	out, _ := cmd.Output()
	if !strings.Contains(string(out), "wrk-merge-back") {
		t.Fatal("old 'wrk-merge-back' stash was unexpectedly consumed")
	}

	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
