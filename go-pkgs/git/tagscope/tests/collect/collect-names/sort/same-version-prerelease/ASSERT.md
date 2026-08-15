## Expected

- Root scope `Tags` order is `v0.0.1`, `v0.0.1-beta`, `v0.0.1-alpha`.
- `Newest.FullName` is `v0.0.1` (numeric release).
- `LatestRelease.FullName` is `v0.0.1`.
- `HasPrereleaseHead` is false.

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
	want := []string{"v0.0.1", "v0.0.1-beta", "v0.0.1-alpha"}
	for i, name := range want {
		if lineage.Tags[i].FullName != name {
			t.Fatalf("Tags[%d] = %q, want %q", i, lineage.Tags[i].FullName, name)
		}
	}
	if lineage.Newest == nil || lineage.Newest.FullName != "v0.0.1" {
		t.Fatalf("Newest = %v, want v0.0.1", lineage.Newest)
	}
	if lineage.LatestRelease == nil || lineage.LatestRelease.FullName != "v0.0.1" {
		t.Fatalf("LatestRelease = %v, want v0.0.1", lineage.LatestRelease)
	}
	if lineage.HasPrereleaseHead {
		t.Fatal("HasPrereleaseHead = true, want false")
	}
}
```