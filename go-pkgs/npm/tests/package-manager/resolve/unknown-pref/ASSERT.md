## Expected

- `resp` is nil.

## Errors

- `err` is non-nil and mentions `yarnberry`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for unknown manager")
	}
	if !strings.Contains(err.Error(), "yarnberry") {
		t.Fatalf("expected yarnberry in error, got %v", err)
	}
}```
