## Expected

- `ApplyDirs(target)` with no dirs returns nil.
- Target directory exists and has no entries other than possibly empty dir listing.

## Errors

- No error.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("ApplyDirs empty: %v", err)
	}
	ents, readErr := os.ReadDir(resp.Target)
	if readErr != nil {
		t.Fatalf("ReadDir target: %v", readErr)
	}
	if len(ents) != 0 {
		names := make([]string, 0, len(ents))
		for _, e := range ents {
			names = append(names, e.Name())
		}
		t.Fatalf("target should be empty after no-op ApplyDirs, got %v", names)
	}
}
```
