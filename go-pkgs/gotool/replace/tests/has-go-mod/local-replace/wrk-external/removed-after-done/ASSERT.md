## Expected

- No error from the scan.
- Exactly 1 issue with `IsIntraRepo == true` (path prefix is still inside the scanning worktree).
- Lenient callers skip this issue (0 blocking).

## Exit Code

- No error.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("CheckLocalReplaces returned error: %v", resp.Err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("expected 1 within-worktree issue after external remove, got %d: %+v", len(resp.Issues), resp.Issues)
	}
	if !resp.Issues[0].IsIntraRepo {
		t.Fatalf("expected IsIntraRepo=true: removed external path still under worktree root, got %+v", resp.Issues[0])
	}
}
```
