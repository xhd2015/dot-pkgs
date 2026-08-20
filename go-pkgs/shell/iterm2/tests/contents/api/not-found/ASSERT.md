## Expected

- Error contains `session not found` and the UUID

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
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "session not found") {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(err.Error(), defaultContentsSessionID) {
		t.Fatalf("error missing id: %v", err)
	}
}
```
