## Expected

- `resp.Tokens` equals `[]string{"aid", "user"}`.

## Errors

- `err` is nil.

```go
import (
	"reflect"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"aid", "user"}
	if !reflect.DeepEqual(resp.Tokens, want) {
		t.Fatalf("Tokens(%q) = %#v, want %#v", req.Query, resp.Tokens, want)
	}
}
```
