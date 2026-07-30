## Expected

- `resp.Alive` is true when inject returns true for pid 50.
- With inject false for pid 51, `Alive` is false.
- With inject true but pid ≤ 0, still false (non-positive short-circuit wins).

## Errors

- `err` is nil.
- Ignoring inject or returning true for pid ≤ 0 is failure.

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
	if !resp.Alive {
		t.Fatal("Alive(50) with inject true want true")
	}

	falseOpts := proc.Options{
		Alive: func(pid int) bool { return false },
	}
	if proc.Alive(51, falseOpts) {
		t.Fatal("Alive(51) with inject false want false")
	}

	// Non-positive always false even if inject would say true.
	trueOpts := proc.Options{
		Alive: func(pid int) bool { return true },
	}
	if proc.Alive(0, trueOpts) {
		t.Fatal("Alive(0) must ignore inject and return false")
	}
	if proc.Alive(-1, trueOpts) {
		t.Fatal("Alive(-1) must ignore inject and return false")
	}
}
```
