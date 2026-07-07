## Expected

- `resp.Formatted` is `"dirty (0 added, 1 changed, 0 renamed, 0 deleted)"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "dirty (0 added, 1 changed, 0 renamed, 0 deleted)"
	if resp.Formatted != want {
		t.Fatalf("formatted = %q, want %q", resp.Formatted, want)
	}
}
```