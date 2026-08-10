## Expected

- `err` is nil.
- Index map has exactly three entries for type-0 ids: `3→0`, `132→1`, `234→2`.
- Non-type-0 id `50` is **not** present in the map.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("BuildUserSpaceIndex: unexpected err: %v", err)
	}
	if resp == nil || resp.IndexMap == nil {
		t.Fatal("expected non-nil IndexMap")
	}
	want := map[uint64]int{3: 0, 132: 1, 234: 2}
	if len(resp.IndexMap) != len(want) {
		t.Fatalf("IndexMap len=%d want %d; got %#v", len(resp.IndexMap), len(want), resp.IndexMap)
	}
	for id, idx := range want {
		got, ok := resp.IndexMap[id]
		if !ok {
			t.Fatalf("IndexMap missing id %d; got %#v", id, resp.IndexMap)
		}
		if got != idx {
			t.Fatalf("IndexMap[%d]=%d want %d; got %#v", id, got, idx, resp.IndexMap)
		}
	}
	if _, ok := resp.IndexMap[50]; ok {
		t.Fatalf("type!=0 id 50 must be omitted; got %#v", resp.IndexMap)
	}
}
```
