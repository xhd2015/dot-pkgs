## Expected

- `resp.Collected.ByScope` has keys for root (`""`) and `sub/`.
- Root lineage `Newest.FullName` is `v0.0.1`.
- `sub/` lineage `Newest.FullName` is `sub/v0.0.2`.
- Both lineages have `LatestRelease` equal to `Newest`.

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
	if len(resp.Collected.ByScope) != 2 {
		t.Fatalf("ByScope len = %d, want 2", len(resp.Collected.ByScope))
	}
	root := lineageFor(t, resp.Collected, "")
	sub := lineageFor(t, resp.Collected, "sub/")
	if root.Newest == nil || root.Newest.FullName != "v0.0.1" {
		t.Fatalf("root Newest = %v, want v0.0.1", root.Newest)
	}
	if sub.Newest == nil || sub.Newest.FullName != "sub/v0.0.2" {
		t.Fatalf("sub Newest = %v, want sub/v0.0.2", sub.Newest)
	}
	if root.LatestRelease == nil || root.LatestRelease.FullName != "v0.0.1" {
		t.Fatalf("root LatestRelease = %v, want v0.0.1", root.LatestRelease)
	}
	if sub.LatestRelease == nil || sub.LatestRelease.FullName != "sub/v0.0.2" {
		t.Fatalf("sub LatestRelease = %v, want sub/v0.0.2", sub.LatestRelease)
	}
}
```