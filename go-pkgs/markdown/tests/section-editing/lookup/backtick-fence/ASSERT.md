## Expected

The fenced `# Fake` line remains in the selected body.

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
	if resp.Found != true {
		t.Fatalf("found = %v, want true", resp.Found)
	}
	assertExact(t, resp.Content, "before\n\x60\x60\x60md\n# Fake\n\x60\x60\x60\nafter\n")
	assertExact(t, resp.Output, "# Target\nbefore\n\x60\x60\x60md\n# Fake\n\x60\x60\x60\nafter\n# Next\nkeep\n")
	assertExact(t, resp.SecondOutput, "# Target\nbefore\n\x60\x60\x60md\n# Fake\n\x60\x60\x60\nafter\n# Next\nkeep\n")
}
```
