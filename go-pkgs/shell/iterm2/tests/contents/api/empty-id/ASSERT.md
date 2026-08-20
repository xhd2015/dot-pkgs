## Expected

- Error mentions session id required
- Exec is not called

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "session id") {
		t.Fatalf("error = %v", err)
	}
	if resp.ExecN != 0 {
		t.Fatalf("exec called %d times", resp.ExecN)
	}
}
```
