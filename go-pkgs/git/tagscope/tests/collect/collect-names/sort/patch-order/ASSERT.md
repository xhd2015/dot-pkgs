## Expected

- Root scope `Tags` order is `v0.0.10`, `v0.0.2`, `v0.0.1`.
- `Newest.FullName` is `v0.0.10`.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	lineage := lineageFor(t, resp.Collected, "")
	want := []string{"v0.0.10", "v0.0.2", "v0.0.1"}
	if len(lineage.Tags) != len(want) {
		t.Fatalf("Tags len = %d, want %d", len(lineage.Tags), len(want))
	}
	for i, name := range want {
		if lineage.Tags[i].FullName != name {
			t.Fatalf("Tags[%d] = %q, want %q", i, lineage.Tags[i].FullName, name)
		}
	}
	if lineage.Newest == nil || lineage.Newest.FullName != "v0.0.10" {
		t.Fatalf("Newest = %v, want v0.0.10", lineage.Newest)
	}
}
```