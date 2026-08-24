## Expected

- NextOK true, Next equals anchor 10:00 UTC

## Side Effects

- None.

## Errors

- None unless noted in Expected.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 8, 24, 10, 0, 0, 0, utc())
	if !resp.NextOK || !resp.Next.Equal(want) {
		t.Fatalf("Next=%v ok=%v, want %v", resp.Next, resp.NextOK, want)
	}
}
```
