## Expected

- `err == nil`
- `PeriodType == "weekly"`
- `UsedPercent == 2`, `RemainingPercent == 98`
- `GetCount == 2`

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "PeriodType", resp.PeriodType, "weekly")
	assertEqual(t, "UsedPercent", resp.UsedPercent, 2)
	assertEqual(t, "RemainingPercent", resp.RemainingPercent, 98)
	assertEqual(t, "GetCount", resp.GetCount, 2)
	wantReset := time.Date(2026, 9, 4, 0, 55, 25, 179446000, time.UTC).Unix()
	assertEqual(t, "ResetUnix", resp.ResetUnix, wantReset)
}
```
