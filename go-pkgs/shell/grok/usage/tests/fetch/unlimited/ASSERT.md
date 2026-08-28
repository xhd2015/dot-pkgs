## Expected

- `err == nil`
- `Source == "billing"`
- `Used == 73`, `MonthlyLimit == 0`
- `UsedPercent == -1`, `RemainingPercent == -1`
- both GETs return uncapped monthly shape (no weekly %)

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
	assertEqual(t, "Used", resp.Used, int64(73))
	assertEqual(t, "MonthlyLimit", resp.MonthlyLimit, int64(0))
	assertEqual(t, "UsedPercent", resp.UsedPercent, -1)
	assertEqual(t, "RemainingPercent", resp.RemainingPercent, -1)
	assertEqual(t, "GetCount", resp.GetCount, 2)
}
```
