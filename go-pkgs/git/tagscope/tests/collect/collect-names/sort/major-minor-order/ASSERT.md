## Expected

- Root scope `Tags` order is `v2.0.0`, `v1.0.0`, `v0.9.9`.
- `Newest.FullName` is `v2.0.0`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	lineage := lineageFor(t, resp.Collected, "")
	want := []string{"v2.0.0", "v1.0.0", "v0.9.9"}
	for i, name := range want {
		if lineage.Tags[i].FullName != name {
			t.Fatalf("Tags[%d] = %q, want %q", i, lineage.Tags[i].FullName, name)
		}
	}
	if lineage.Newest == nil || lineage.Newest.FullName != "v2.0.0" {
		t.Fatalf("Newest = %v, want v2.0.0", lineage.Newest)
	}
}
```