## Expected

- err mentions every-

## Side Effects

- None.

## Errors

- None unless noted in Expected.

## Exit Code

- N/A (in-process library).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "every-") {
		t.Fatalf("error %q should mention every-", err)
	}
}
```
