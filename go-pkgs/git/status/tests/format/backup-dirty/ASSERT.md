## Expected

- `resp.Formatted` is `"dirty (2 modified, 1 untracked)"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "dirty (2 modified, 1 untracked)"
	if resp.Formatted != want {
		t.Fatalf("formatted = %q, want %q", resp.Formatted, want)
	}
}
```
