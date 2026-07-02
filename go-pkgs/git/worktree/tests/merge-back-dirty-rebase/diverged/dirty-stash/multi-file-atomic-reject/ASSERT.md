## Expected

- `MergeBack` returns error about dirty changes conflicting (README.md conflicts).
- ALL files retain user's content — none were partially applied.
- Source branch NOT updated.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected conflict error for multi-file")
	}
	if !strings.Contains(err.Error(), "dirty changes conflict with rebase") {
		t.Fatalf("expected conflict error, got: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(req.SourcePath, "README.md"))
	if string(content) != "# USER CONFLICT\n" {
		t.Fatalf("README.md was changed, got %q", string(content))
	}
	content1, _ := os.ReadFile(filepath.Join(req.SourcePath, "other-1.txt"))
	if string(content1) != "USER other 1\n" {
		t.Fatalf("other-1.txt was changed, got %q", string(content1))
	}
	content2, _ := os.ReadFile(filepath.Join(req.SourcePath, "other-2.txt"))
	if string(content2) != "USER other 2\n" {
		t.Fatalf("other-2.txt was changed, got %q", string(content2))
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
