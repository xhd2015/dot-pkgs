## Expected

Lookup starts in the middle and the H1 boundary ends the H3 section.

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
	assertExact(t, resp.Content, "value\n")
	assertExact(t, resp.Output, "preamble\n# Parent\n## Area\n### Target\nvalue\n# Final\nend\n")
	assertExact(t, resp.SecondOutput, "preamble\n# Parent\n## Area\n### Target\nvalue\n# Final\nend\n")
}
```

