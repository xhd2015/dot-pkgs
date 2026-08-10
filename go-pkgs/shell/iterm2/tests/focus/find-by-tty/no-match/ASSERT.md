## Expected

- `FindByTTY` returns empty slice (len 0).
- No error from Run.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Refs) != 0 {
		t.Fatalf("FindByTTY no-match: len=%d, want 0; refs=%+v", len(resp.Refs), resp.Refs)
	}
}
```
