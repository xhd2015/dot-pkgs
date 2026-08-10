## Expected

- Exactly one match: SessionID `s-norm`, WindowID `win-norm`.
- Bare query `ttys148` matches ref `/dev/ttys148` via NormalizeTTY.

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
	if len(resp.Refs) != 1 {
		t.Fatalf("FindByTTY normalization: len=%d, want 1; refs=%+v", len(resp.Refs), resp.Refs)
	}
	r := resp.Refs[0]
	if r.SessionID != "s-norm" || r.WindowID != "win-norm" {
		t.Fatalf("match = %+v, want s-norm / win-norm", r)
	}
	if r.TTY != "/dev/ttys148" {
		t.Fatalf("matched ref TTY = %q, want /dev/ttys148", r.TTY)
	}
}
```
