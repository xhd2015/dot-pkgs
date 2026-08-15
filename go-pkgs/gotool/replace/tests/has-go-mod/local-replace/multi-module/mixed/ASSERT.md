## Expected

- No error from the scan.
- Exactly 2 issues are returned.
- One issue has `IsIntraRepo == true` (root go.mod, `./sub`).
- One issue has `IsIntraRepo == false` (sub2/go.mod, `./nonexistent`).

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
	if len(resp.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %+v", len(resp.Issues), resp.Issues)
	}
	hasIntra := false
	hasExtra := false
	for _, issue := range resp.Issues {
		if issue.IsIntraRepo {
			hasIntra = true
		} else {
			hasExtra = true
		}
	}
	if !hasIntra {
		t.Fatalf("expected at least one intra-repo issue, got none: %+v", resp.Issues)
	}
	if !hasExtra {
		t.Fatalf("expected at least one extra-repo issue, got none: %+v", resp.Issues)
	}
}
```