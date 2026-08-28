## Expected

- `err == nil`
- `GetCount == 3` (monthly 401, monthly retry, credits)
- second Ensure has `ForceRefresh=true`
- `UsedPercent == 73` (monthly preferred)

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
	assertEqual(t, "GetCount", resp.GetCount, 3)
	assertEqual(t, "UsedPercent", resp.UsedPercent, 73)
	if len(resp.EnsureForced) != 2 || resp.EnsureForced[1] != true {
		t.Fatalf("EnsureForced = %v", resp.EnsureForced)
	}
}
```