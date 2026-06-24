## Expected

- `resp.ParseOK` is false.
- `resp.Owner` and `resp.Repo` are empty.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ParseOK {
		t.Fatal("expected ParseOK false for unparseable URL")
	}
	if resp.Owner != "" || resp.Repo != "" {
		t.Fatalf("expected empty owner/repo, got %q/%q", resp.Owner, resp.Repo)
	}
}
```