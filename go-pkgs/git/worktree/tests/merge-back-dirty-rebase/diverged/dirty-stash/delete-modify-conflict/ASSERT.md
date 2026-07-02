## Expected

- `MergeBack` returns error about dirty changes conflicting.
- `README.md` still absent from source working tree.
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
		t.Fatal("expected delete/modify conflict error")
	}
	if !strings.Contains(err.Error(), "dirty changes conflict with rebase") {
		t.Fatalf("expected conflict error, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(req.SourcePath, "README.md")); err == nil {
		t.Fatal("README.md was restored — user's deletion was lost")
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
}
```
