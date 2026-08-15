## Expected

- `err` is nil.
- `resp.IsBinary == false` (file is Unicode/UTF-8 text with box-drawing characters).
- `resp.Desc` is text-like (e.g. `"text"` or `"text file"`), **not** `"binary file"`.

## Errors

- Classifying the snapshot as binary (current buggy RED path).
- Non-nil error opening/reading the fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("DetectFileType(%q): %v", req.Path, err)
	}
	if resp.IsBinary {
		t.Fatalf("DetectFileType(%q) isBinary=true desc=%q, want isBinary=false for UTF-8 TTY snapshot",
			req.Path, resp.Desc)
	}
	if resp.Desc == "binary file" {
		t.Fatalf("DetectFileType(%q) desc=%q, want text-like (not \"binary file\")",
			req.Path, resp.Desc)
	}
}
```
