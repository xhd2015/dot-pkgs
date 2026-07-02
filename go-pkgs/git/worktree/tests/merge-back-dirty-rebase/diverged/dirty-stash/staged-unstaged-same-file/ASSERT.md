## Expected

- `MergeBack` succeeds, action `"rebased-and-merged"`.
- `multi.txt` has user's full content (both staged and unstaged changes reflected).
- Note: staged/unstaged distinction may be lost (everything comes back as unstaged).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	content, _ := os.ReadFile(filepath.Join(req.SourcePath, "multi.txt"))
	if !strings.Contains(string(content), "LINE TWO STAGED") {
		t.Fatalf("staged change lost: %q", string(content))
	}
	if !strings.Contains(string(content), "LINE THREE UNSTAGED") {
		t.Fatalf("unstaged change lost: %q", string(content))
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
