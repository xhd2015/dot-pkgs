## Expected

- Two matches: A then C (stable order of input refs).
- Not C then A (must not reorder by query list).
- SessionIDs `sA`, `sC`.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Refs) != 2 {
		t.Fatalf("FindByTTY multi-input: len=%d, want 2; refs=%+v", len(resp.Refs), resp.Refs)
	}
	if resp.Refs[0].SessionID != "sA" || resp.Refs[0].Name != "A" {
		t.Fatalf("refs[0] = %+v, want sA/A (stable ref order)", resp.Refs[0])
	}
	if resp.Refs[1].SessionID != "sC" || resp.Refs[1].Name != "C" {
		t.Fatalf("refs[1] = %+v, want sC/C", resp.Refs[1])
	}
}
```
