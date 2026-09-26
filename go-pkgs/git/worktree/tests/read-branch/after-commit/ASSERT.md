## Expected

- `ReadBranch` returns `"main"` after the first commit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("ReadBranch after commit: %v", err)
	}
	if resp.Branch != "main" {
		t.Fatalf("Branch = %q, want main", resp.Branch)
	}
}
```
