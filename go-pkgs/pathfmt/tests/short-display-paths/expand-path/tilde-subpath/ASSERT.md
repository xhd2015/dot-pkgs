## Expected

- `resp.Display` equals `filepath.Join(home, "foo", "bar")`.

## Errors

- `err` is nil.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "foo", "bar")
	if resp.Display != want {
		t.Fatalf("expected %q, got %q", want, resp.Display)
	}
}```
