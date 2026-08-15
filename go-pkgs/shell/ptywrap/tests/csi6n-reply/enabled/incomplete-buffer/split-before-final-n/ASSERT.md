## Expected

- `err` is nil.
- `resp.Replies` is `\x1b[3;7R`.
- `resp.Rest` is empty.
- WriteCalls ≥ 1.

## Errors

- None.

```go
import (
	"bytes"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("\x1b[3;7R")
	if !bytes.Equal(resp.Replies, want) {
		t.Fatalf("replies=%q want %q (WriteCalls=%d Rest=%q)", resp.Replies, want, resp.WriteCalls, resp.Rest)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("rest should be empty after complete, got %q", resp.Rest)
	}
	if resp.WriteCalls < 1 {
		t.Fatalf("expected write after second chunk, WriteCalls=%d", resp.WriteCalls)
	}
}
```
