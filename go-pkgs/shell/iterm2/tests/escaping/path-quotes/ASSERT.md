## Expected

- Inner quotes are backslash-escaped for AppleScript literals.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := `/tmp/\"proj\"`
	if resp.Escaped != want {
		t.Fatalf("Escaped = %q, want %q", resp.Escaped, want)
	}
}
```