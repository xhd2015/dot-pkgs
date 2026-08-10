## Expected

- `resp.Display` is `"a/b/c"` (platform-native separators).

## Errors

- `err` is nil.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "a" + string(filepath.Separator) + "b" + string(filepath.Separator) + "c"
	if resp.Display != want {
		t.Fatalf("expected %q, got %q", want, resp.Display)
	}
}```
