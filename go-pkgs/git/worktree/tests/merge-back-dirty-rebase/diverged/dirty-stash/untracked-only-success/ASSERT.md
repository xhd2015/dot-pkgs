## Expected

- `MergeBack` succeeds, action `"rebased-and-merged"`.
- Both untracked files survive in source.
- Source branch merged into main.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success for untracked-only dirt, got: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	for _, name := range []string{"untracked-1.txt", "untracked-2.txt"} {
		content, err := os.ReadFile(filepath.Join(req.SourcePath, name))
		if err != nil {
			t.Fatalf("%s lost: %v", name, err)
		}
		if len(content) == 0 {
			t.Fatalf("%s is empty after merge-back", name)
		}
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
