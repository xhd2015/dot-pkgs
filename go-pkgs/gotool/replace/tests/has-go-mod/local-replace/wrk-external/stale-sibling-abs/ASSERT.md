## Expected

- No error from the scan.
- Exactly 1 issue with `IsIntraRepo == false` (target is outside the scanning worktree).

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
		t.Fatalf("expected 1 outside-worktree issue, got %d: %+v", len(resp.Issues), resp.Issues)
	}
	issue := resp.Issues[0]
	if issue.IsIntraRepo {
		t.Fatalf("expected IsIntraRepo=false for sibling worktree path, got true: %+v", issue)
	}
}

```