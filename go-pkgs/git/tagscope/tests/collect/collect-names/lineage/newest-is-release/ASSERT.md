## Expected

- Root scope lineage `Tags` order is `v0.0.2`, `v0.0.1`.
- `Newest.FullName` is `v0.0.2`.
- `LatestRelease.FullName` is `v0.0.2`.
- `HasPrereleaseHead` is false.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	lineage := lineageFor(t, resp.Collected, "")
	if len(lineage.Tags) != 2 || lineage.Tags[0].FullName != "v0.0.2" || lineage.Tags[1].FullName != "v0.0.1" {
		t.Fatalf("Tags = %+v, want newest-first v0.0.2, v0.0.1", lineage.Tags)
	}
	if lineage.Newest == nil || lineage.Newest.FullName != "v0.0.2" {
		t.Fatalf("Newest = %v, want v0.0.2", lineage.Newest)
	}
	if lineage.LatestRelease == nil || lineage.LatestRelease.FullName != "v0.0.2" {
		t.Fatalf("LatestRelease = %v, want v0.0.2", lineage.LatestRelease)
	}
	if lineage.HasPrereleaseHead {
		t.Fatal("HasPrereleaseHead = true, want false")
	}
}
```