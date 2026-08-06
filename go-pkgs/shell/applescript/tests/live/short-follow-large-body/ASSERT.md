---
label: e2e
---

## Expected

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.LiveSkipped {
		t.Skip(resp.LiveSkipReason)
	}
	if !resp.LiveMatch {
		t.Fatalf("short-follow large body: want exact match gotLen=%d wantLen=%d",
			resp.LiveGotLen, resp.LiveWantLen)
	}
	if resp.LiveWantLen < 2000 {
		t.Fatalf("fixture should be multi-KB wantLen=%d", resp.LiveWantLen)
	}
}
```
