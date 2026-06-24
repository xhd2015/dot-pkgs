## Expected

- `resp.FullName` is `o/r`.

## Side Effects

- None (string construction).

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.FullName != "o/r" {
		t.Fatalf("expected FullName o/r, got %q", resp.FullName)
	}
}```