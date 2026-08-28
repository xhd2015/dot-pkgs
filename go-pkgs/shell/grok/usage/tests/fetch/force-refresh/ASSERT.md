## Expected

- `err == nil`
- first Ensure call has `ForceRefresh=true`
- `Source == "billing"`

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Source", resp.Source, "billing")
	if len(resp.EnsureForced) < 1 || !resp.EnsureForced[0] {
		t.Fatalf("EnsureForced = %v want first true", resp.EnsureForced)
	}
}
```
