## Expected

- No error from the scan.
- Exactly 1 issue is returned.
- The issue has `IsIntraRepo == true`.
- The issue has `NewPath == "./sub"`.

```go
import "github.com/xhd2015/doctest/session"

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
	if issue.NewPath != "./sub" {
		t.Fatalf("expected NewPath=./sub, got %q", issue.NewPath)
	}
}
```