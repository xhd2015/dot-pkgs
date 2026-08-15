## Expected

- No error from the scan.
- Exactly 1 issue is returned.
- The issue has `IsIntraRepo == true`.
- The issue has `NewPath` equal to the absolute path.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("expected no error, got: %v", resp.Err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %+v", len(resp.Issues), resp.Issues)
	}
	issue := resp.Issues[0]
	if !issue.IsIntraRepo {
		t.Fatalf("expected IsIntraRepo=true, got false. issue: %+v", issue)
	}
	expectedPath := filepath.Join(req.RootDir, "internal", "pkg")
	if issue.NewPath != expectedPath {
		t.Fatalf("expected NewPath=%q, got %q", expectedPath, issue.NewPath)
	}
}
```