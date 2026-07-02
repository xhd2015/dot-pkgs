## Expected

- No error from the scan.
- Exactly 1 issue is returned.
- The issue has `IsIntraRepo == false`.
- The issue has `NewPath` equal to the external absolute path.

```go
import "path/filepath"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	if issue.IsIntraRepo {
		t.Fatalf("expected IsIntraRepo=false, got true. issue: %+v", issue)
	}
	if issue.NewPath == "" {
		t.Fatal("expected non-empty NewPath")
	}
	// NewPath should be an absolute path outside the repo
	if !filepath.IsAbs(issue.NewPath) {
		t.Fatalf("expected absolute path, got %q", issue.NewPath)
	}
}
```