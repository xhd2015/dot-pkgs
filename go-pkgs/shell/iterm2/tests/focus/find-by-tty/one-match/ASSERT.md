## Expected

- Exactly one match.
- WindowID `win-2`, TabIndex `3`, SessionID `s2`, TTY `/dev/ttys149`, Name `B`.

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
		t.Fatalf("FindByTTY one-match: len=%d, want 1; refs=%+v", len(resp.Refs), resp.Refs)
	}
	r := resp.Refs[0]
	if r.WindowID != "win-2" || r.TabIndex != 3 || r.SessionID != "s2" {
		t.Fatalf("match identity = %+v, want win-2 tab 3 s2", r)
	}
	if r.TTY != "/dev/ttys149" || r.Name != "B" {
		t.Fatalf("match tty/name = %q/%q", r.TTY, r.Name)
	}
}
```
