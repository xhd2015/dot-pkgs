## Expected

- No error.
- `Args` =
  `["open", "-na", "Firefox", "--args", "-new-window", "http://127.0.0.1:8422/projects/eluc"]`.
- The switch is `-new-window` (one dash), not Chromium's `--new-window`.

## Side Effects

- None: `BrowserArgs` is pure and never launches.

## Errors

- None for a non-empty URL with a known alias on darwin.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertArgs(t, resp.Args, []string{
		"open", "-na", "Firefox", "--args", "-new-window",
		"http://127.0.0.1:8422/projects/eluc",
	})
	assertNoRun(t, resp)
}
```
