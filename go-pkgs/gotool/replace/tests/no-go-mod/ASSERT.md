## Expected

- No error is returned.
- The issues slice is nil or empty.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("expected no error, got: %v", resp.Err)
	}
	if len(resp.Issues) != 0 {
		t.Fatalf("expected 0 issues, got %d: %+v", len(resp.Issues), resp.Issues)
	}
}
```