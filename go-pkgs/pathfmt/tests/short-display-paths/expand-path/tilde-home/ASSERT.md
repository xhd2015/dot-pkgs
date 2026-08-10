## Expected

- `resp.Display` equals `os.UserHomeDir()` (absolute home path).

## Errors

- `err` is nil.

```go
import (
	"github.com/xhd2015/doctest/session"
	"os"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Display != home {
		t.Fatalf("expected home %q, got %q", home, resp.Display)
	}
}```
