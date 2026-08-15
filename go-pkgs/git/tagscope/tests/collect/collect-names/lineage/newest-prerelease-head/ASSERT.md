## Expected

- Root scope lineage `Tags` order is `v0.0.3-alpha`, `v0.0.2`.
- `Newest.FullName` is `v0.0.3-alpha`.
- `LatestRelease.FullName` is `v0.0.2`.
- `HasPrereleaseHead` is true.

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
	if len(lineage.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(lineage.Tags))
	}
	if lineage.Tags[0].FullName != "v0.0.3-alpha" || lineage.Tags[1].FullName != "v0.0.2" {
		t.Fatalf("Tags order = [%s, %s], want [v0.0.3-alpha, v0.0.2]", lineage.Tags[0].FullName, lineage.Tags[1].FullName)
	}
	if lineage.Newest == nil || lineage.Newest.FullName != "v0.0.3-alpha" {
		t.Fatalf("Newest = %v, want v0.0.3-alpha", lineage.Newest)
	}
	if lineage.LatestRelease == nil || lineage.LatestRelease.FullName != "v0.0.2" {
		t.Fatalf("LatestRelease = %v, want v0.0.2", lineage.LatestRelease)
	}
	if !lineage.HasPrereleaseHead {
		t.Fatal("HasPrereleaseHead = false, want true")
	}
}
```