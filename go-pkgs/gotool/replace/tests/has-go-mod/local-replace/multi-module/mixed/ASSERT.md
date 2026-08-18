## Expected

- No error from the scan.
- Exactly 2 issues are returned.
- Both issues have `IsIntraRepo == true` (paths resolve under the scan root, including missing `./nonexistent`).

```go
import "github.com/xhd2015/doctest/session"

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
	for _, issue := range resp.Issues {
		if !issue.IsIntraRepo {
			t.Fatalf("expected all issues IsIntraRepo=true, got extra-repo: %+v", issue)
		}
	}
}
```