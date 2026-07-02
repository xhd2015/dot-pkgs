## Expected

- No error from the scan.
- Exactly 1 issue is returned.
- The issue is from the `sub/go.mod` module.
- The issue has `IsIntraRepo == true` (target is inside repo).

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
	if !issue.IsIntraRepo {
		t.Fatalf("expected IsIntraRepo=true, got false. issue: %+v", issue)
	}
	if issue.NewPath != "./local" {
		t.Fatalf("expected NewPath=./local, got %q", issue.NewPath)
	}
	// GoModPath should be sub/go.mod
	expectedGoMod := filepath.Join(req.RootDir, "sub", "go.mod")
	if issue.GoModPath != expectedGoMod {
		t.Fatalf("expected GoModPath=%q, got %q", expectedGoMod, issue.GoModPath)
	}
}
```