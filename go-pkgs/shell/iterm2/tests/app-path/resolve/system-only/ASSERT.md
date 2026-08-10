## Expected

- Resolved path is `/Applications/iTerm.app`.

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
	if resp.ResolvedPath != systemApp {
		t.Fatalf("ResolveAppPathWith() = %q, want %q", resp.ResolvedPath, systemApp)
	}
}
```
