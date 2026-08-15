## Expected

- The captured log contains a line with "via upstream proxy"

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Output, "via upstream proxy") {
		t.Fatalf("expected 'via upstream proxy' in output, got:\n%s", resp.Output)
	}
}
```
