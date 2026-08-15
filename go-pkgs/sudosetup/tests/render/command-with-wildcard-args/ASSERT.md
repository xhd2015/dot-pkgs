## Expected

- `RenderSudoersLine` returns:
  `testuser ALL=(root) NOPASSWD: /opt/homebrew/bin/sing-box run -c *`

## Side Effects

- None (pure render).

## Errors

- None.

## Exit Code

- Success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	want := "testuser ALL=(root) NOPASSWD: /opt/homebrew/bin/sing-box run -c *"
	assertEqual(t, "RenderedLine", resp.RenderedLine, want)
}
```