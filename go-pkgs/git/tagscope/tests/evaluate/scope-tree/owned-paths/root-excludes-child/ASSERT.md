## Expected

- `OwnedPaths` is `["README", "go.mod"]` (excludes `sub/` prefix paths).

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"README", "go.mod"}
	if len(resp.OwnedPaths) != len(want) {
		t.Fatalf("OwnedPaths = %v, want %v", resp.OwnedPaths, want)
	}
	for i := range want {
		if resp.OwnedPaths[i] != want[i] {
			t.Fatalf("OwnedPaths = %v, want %v", resp.OwnedPaths, want)
		}
	}
}
```