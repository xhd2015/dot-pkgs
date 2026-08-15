## Expected

- `MergeBack` returns error about dirty changes conflicting with rebase.
- Source README.md still has user's content.
- Source branch was NOT updated.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected conflict error for content modify/modify")
	}
	if !strings.Contains(err.Error(), "dirty changes conflict with rebase") {
		t.Fatalf("expected conflict error, got: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(req.SourcePath, "README.md"))
	if string(content) != "# USER CHANGED SAME LINE\n" {
		t.Fatalf("README.md was overwritten, got %q", string(content))
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
