## Expected

- `resp.Alive` is false for pid `0`.
- `proc.Alive(-3, proc.Options{})` is also false (direct probe in Assert).

## Errors

- `err` is nil.
- True for non-positive pid is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/proc"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.Alive {
		t.Fatal("Alive(0) want false")
	}
	if proc.Alive(-3, proc.Options{}) {
		t.Fatal("Alive(-3) want false")
	}
}
```
