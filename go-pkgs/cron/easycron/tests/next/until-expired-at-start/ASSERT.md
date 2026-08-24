## Expected

- NextOK false

## Side Effects

- None.

## Errors

- None unless noted in Expected.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.NextOK {
		t.Fatalf("expected expired, got %v", resp.Next)
	}
}
```
