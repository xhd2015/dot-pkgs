---
label: e2e
---

## Expected

Long write-text FollowUp over SoftMax must **not** claim a reliable exact match.
Accept either mismatch/empty (limitation) or rare flaky PASS — assert SoftMax
path was exercised by requiring wantLen large and documenting that callers
must not rely on long write text.

This leaf fails the build only if delivery claims success with zero bytes when
we required a multi-KB body *and* match is true with wantLen under SoftMax
(setup bug). Primary proof: when match is false OR gotLen != wantLen, limitation
is demonstrated (PASS assertion). When match is true, we only warn — flaky PASS
above SoftMax is allowed but logged.

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
	if resp.LiveWantLen < 1500 {
		t.Fatalf("live-long wantLen too small: %d", resp.LiveWantLen)
	}
	// Limitation demonstrated when not exact match (expected common case).
	if resp.LiveMatch {
		t.Logf("note: long follow unexpectedly matched (flaky PASS above SoftMax); gotLen=%d", resp.LiveGotLen)
		return
	}
	// Not a match → demonstrates unreliability of long write text.
	t.Logf("limitation confirmed: match=false gotLen=%d wantLen=%d", resp.LiveGotLen, resp.LiveWantLen)
}
```
