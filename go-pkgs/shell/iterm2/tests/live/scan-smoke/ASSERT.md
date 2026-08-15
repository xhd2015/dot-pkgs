---
label: side-effect-open-iterm2
explanation: >-
  Runs real osascript against iTerm2. May focus or touch windows. Skipped unless
  doctest test --label side-effect-open-iterm2 is used.
---

## Expected

- `osascript` smoke returns stdout `ok`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("live smoke failed: %v", err)
	}
	if resp.LiveStdout != "ok" {
		t.Fatalf("stdout=%q, want ok", resp.LiveStdout)
	}
}
```