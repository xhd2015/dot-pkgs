## Expected

- Resolved path is empty string.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ResolvedPath != "" {
		t.Fatalf("ResolveAppPathWith() = %q, want empty", resp.ResolvedPath)
	}
}
```
