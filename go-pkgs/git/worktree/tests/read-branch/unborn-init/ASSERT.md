## Expected

- `ReadBranch` returns `"main"`.
- No error (must not fatal on unborn `HEAD`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("ReadBranch on unborn HEAD: %v", err)
	}
	if resp.Branch != "main" {
		t.Fatalf("Branch = %q, want main", resp.Branch)
	}
}
```
