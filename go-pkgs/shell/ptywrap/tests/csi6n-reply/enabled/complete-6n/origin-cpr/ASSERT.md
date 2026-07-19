## Expected

- `err` is nil.
- `resp.Replies` is exactly `\x1b[1;1R`.
- `resp.Rest` is empty.

## Errors

- None.

```go
import (
	"bytes"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("\x1b[1;1R")
	if !bytes.Equal(resp.Replies, want) {
		t.Fatalf("replies=%q want %q", resp.Replies, want)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("rest should be empty, got %q", resp.Rest)
	}
}
```
