## Expected

- No error.
- Exactly one runner call with
  `["open", "-na", "Opera", "--args", "--new-window", "http://127.0.0.1:8422/projects/eluc"]`.
- `Result.Args` is that argv and `Out` is `"launched"`.
- The document form `{"open", "-n", url}` is **not** used: it would open a tab in
  the browser's existing window.

## Side Effects

- Exactly one runner call. The real default browser is never consulted: the
  resolver is injected through `Config.DefaultBrowser`.

## Errors

- None: the target is valid, the platform is darwin, and the default resolves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	want := []string{
		"open", "-na", "Opera", "--args", "--new-window",
		"http://127.0.0.1:8422/projects/eluc",
	}
	assertRunOnce(t, resp, want)
	assertArgs(t, resp.Args, want)
	assertOut(t, resp, "launched")
}
```
