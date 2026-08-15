## Expected

- Backslash and double-quote are escaped for AppleScript string literals.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := `say \"hi\"\\x`
	if resp.Escaped != want {
		t.Fatalf("Escaped = %q, want %q", resp.Escaped, want)
	}
}
```
