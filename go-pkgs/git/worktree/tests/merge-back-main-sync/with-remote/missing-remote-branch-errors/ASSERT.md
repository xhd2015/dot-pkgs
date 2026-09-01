# Assert

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}
	if resp.Err == "" {
		t.Fatal("expected fetch error, got success")
	}
	if !strings.Contains(resp.Err, "main-sync") {
		t.Fatalf("error %q should mention main-sync", resp.Err)
	}
}
```
