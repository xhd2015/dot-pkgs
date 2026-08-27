## Expected

- `err == nil`
- `Source == "wham/usage"`
- `RemainingPercent == 38`
- two GetURLs: first `codex/usage`, second `wham/usage`

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
	assertEqual(t, "Source", resp.Source, "wham/usage")
	assertEqual(t, "RemainingPercent", resp.RemainingPercent, 38)
	if len(resp.GetURLs) != 2 {
		t.Fatalf("GetURLs = %v", resp.GetURLs)
	}
	assertContains(t, "url0", resp.GetURLs[0], "codex/usage")
	assertContains(t, "url1", resp.GetURLs[1], "wham/usage")
}
```
