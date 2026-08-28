## Expected

- `err == nil`
- `Source == "billing"`
- `Used == 73`, `MonthlyLimit == 100`
- `UsedPercent == 73`, `RemainingPercent == 27`
- `Email == "user@example.com"`
- `ResetUnix == 1788220800` (2026-09-01 UTC)
- one GET to billing URL; Ensure ForceRefresh false

```go
import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Source", resp.Source, "billing")
	assertEqual(t, "Used", resp.Used, int64(73))
	assertEqual(t, "MonthlyLimit", resp.MonthlyLimit, int64(100))
	assertEqual(t, "UsedPercent", resp.UsedPercent, 73)
	assertEqual(t, "RemainingPercent", resp.RemainingPercent, 27)
	assertEqual(t, "Email", resp.Email, "user@example.com")
	wantReset := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Unix()
	assertEqual(t, "ResetUnix", resp.ResetUnix, wantReset)
	assertEqual(t, "GetCount", resp.GetCount, 1)
	if len(resp.GetURLs) != 1 || !strings.Contains(resp.GetURLs[0], "/v1/billing") {
		t.Fatalf("GetURLs = %v", resp.GetURLs)
	}
	if len(resp.EnsureForced) != 1 || resp.EnsureForced[0] {
		t.Fatalf("EnsureForced = %v", resp.EnsureForced)
	}
}
```
