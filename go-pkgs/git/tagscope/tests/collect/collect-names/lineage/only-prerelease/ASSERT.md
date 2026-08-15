## Expected

- Root scope lineage has one tag `v0.0.1-alpha`.
- `Newest.FullName` is `v0.0.1-alpha`.
- `LatestRelease` is nil.
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
	if len(lineage.Tags) != 1 || lineage.Tags[0].FullName != "v0.0.1-alpha" {
		t.Fatalf("Tags = %+v, want [v0.0.1-alpha]", lineage.Tags)
	}
	if lineage.Newest == nil || lineage.Newest.FullName != "v0.0.1-alpha" {
		t.Fatalf("Newest = %v, want v0.0.1-alpha", lineage.Newest)
	}
	if lineage.LatestRelease != nil {
		t.Fatalf("LatestRelease = %v, want nil", lineage.LatestRelease)
	}
	if !lineage.HasPrereleaseHead {
		t.Fatal("HasPrereleaseHead = false, want true")
	}
}
```