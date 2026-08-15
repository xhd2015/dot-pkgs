## Expected

- `resp.ParseOK` is true.
- `resp.Owner` is `"xhd2015"`.
- `resp.Repo` is `"lifelog"`.

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
		t.Fatal("expected ParseOK true for GitHub SSH URL")
	}
	if resp.Owner != "xhd2015" {
		t.Fatalf("owner = %q, want xhd2015", resp.Owner)
	}
	if resp.Repo != "lifelog" {
		t.Fatalf("repo = %q, want lifelog", resp.Repo)
	}
}
```