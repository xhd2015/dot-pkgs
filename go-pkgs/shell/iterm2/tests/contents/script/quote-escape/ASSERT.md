## Expected

- Script contains escaped `\"` for the quote in the session id.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Script, `aa\"bb`) {
		t.Fatalf("expected escaped quote in script:\n%s", resp.Script)
	}
}
```
