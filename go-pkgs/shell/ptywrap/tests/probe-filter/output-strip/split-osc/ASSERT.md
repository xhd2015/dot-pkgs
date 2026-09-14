## Expected

- `err` is nil.
- Combined `resp.Out` is `himore`.
- Final `resp.Rest` is empty.

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
	want := []byte("himore")
	if !bytes.Equal(resp.Out, want) {
		t.Fatalf("out=%q want %q", resp.Out, want)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("rest=%q", resp.Rest)
	}
}
```
