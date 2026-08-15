## Expected

- `MergeBack` returns error about dirty changes conflicting.
- `new.txt` in source still has user's content.
- Source branch NOT updated.

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
		t.Fatal("expected add/add conflict error")
	}
	if !strings.Contains(err.Error(), "dirty changes conflict with rebase") {
		t.Fatalf("expected conflict error, got: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(req.SourcePath, "new.txt"))
	if string(content) != "USER untracked\n" {
		t.Fatalf("new.txt content was changed, got %q", string(content))
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
