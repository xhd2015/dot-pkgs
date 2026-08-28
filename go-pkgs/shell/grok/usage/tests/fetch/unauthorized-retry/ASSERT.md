## Expected

- `err == nil`
- `GetCount == 2`
- second Ensure has `ForceRefresh=true`
- `UsedPercent == 73`

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
	assertEqual(t, "GetCount", resp.GetCount, 2)
	assertEqual(t, "UsedPercent", resp.UsedPercent, 73)
	if len(resp.EnsureForced) != 2 || resp.EnsureForced[1] != true {
		t.Fatalf("EnsureForced = %v", resp.EnsureForced)
	}
}
```
