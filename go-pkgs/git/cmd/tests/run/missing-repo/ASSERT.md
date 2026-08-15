## Expected

- `resp` is nil.
- `err` is non-nil and mentions git failure.

## Errors

- Run must not succeed for a directory outside any git work tree.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error running git in non-repo directory")
	}
	if resp != nil {
		t.Fatalf("expected nil response on error, got %+v", resp)
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "git") {
		t.Fatalf("error should mention git, got: %v", err)
	}
}
```
