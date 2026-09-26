## Expected

- `ReadBranch` returns `"HEAD"` on detached HEAD (unchanged contract).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("ReadBranch on detached HEAD: %v", err)
	}
	if resp.Branch != "HEAD" {
		t.Fatalf("Branch = %q, want HEAD", resp.Branch)
	}
}
```
