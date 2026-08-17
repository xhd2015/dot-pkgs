## Expected

The body includes the child section and stops immediately before `# Beta`.

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
	assertExact(t, resp.Content, "body\n## Child\nchild body\n")
	assertExact(t, resp.Output, "# Alpha\nbody\n## Child\nchild body\n# Beta\nkeep\n")
	assertExact(t, resp.SecondOutput, "# Alpha\nbody\n## Child\nchild body\n# Beta\nkeep\n")
}
```

