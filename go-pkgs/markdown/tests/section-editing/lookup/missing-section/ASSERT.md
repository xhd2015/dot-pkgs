## Expected

`found` is false, content is empty, and no error is returned.

## Errors

- The returned error and sentinel identity must match this scenario.
- Exact text comparisons include every byte and line ending.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/markdown"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Found != false {
		t.Fatalf("found = %v, want false", resp.Found)
	}
	assertExact(t, resp.Content, "")
	assertExact(t, resp.Output, "# Existing\n")
	assertExact(t, resp.SecondOutput, "# Existing\n")
}
```

