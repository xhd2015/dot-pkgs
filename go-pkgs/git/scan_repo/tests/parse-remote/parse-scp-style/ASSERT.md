## Expected

- `resp.ParseOK` is true.
- `resp.Owner` is `"acme"`.
- `resp.Repo` is `"widget"`.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.ParseOK {
		t.Fatal("expected ParseOK true for SCP-style URL")
	}
	if resp.Owner != "acme" {
		t.Fatalf("owner = %q, want acme", resp.Owner)
	}
	if resp.Repo != "widget" {
		t.Fatalf("repo = %q, want widget", resp.Repo)
	}
}
```