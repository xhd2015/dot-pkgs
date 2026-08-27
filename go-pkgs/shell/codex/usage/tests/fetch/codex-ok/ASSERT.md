## Expected

- `err == nil`
- `Source == "codex/usage"`
- `RemainingPercent == 38`, `UsedPercent == 62`
- `Email == "user@example.com"`
- `PlanType == "business"`
- `ResetUnix == 1788220800`
- exactly one GetURL containing `codex/usage`

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Source", resp.Source, "codex/usage")
	assertEqual(t, "RemainingPercent", resp.RemainingPercent, 38)
	assertEqual(t, "UsedPercent", resp.UsedPercent, 62)
	assertEqual(t, "Email", resp.Email, "user@example.com")
	assertEqual(t, "PlanType", resp.PlanType, "business")
	assertEqual(t, "ResetUnix", resp.ResetUnix, int64(1788220800))
	if len(resp.GetURLs) != 1 || !strings.Contains(resp.GetURLs[0], "codex/usage") {
		t.Fatalf("GetURLs = %v", resp.GetURLs)
	}
}
```
