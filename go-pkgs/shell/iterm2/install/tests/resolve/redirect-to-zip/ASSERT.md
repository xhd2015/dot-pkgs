## Expected

- `err == nil`.
- `resp.Version == "3.6.11"`.
- `resp.URL` contains `iTerm2-3_6_11.zip`.
- `resp.URL` does not contain `arm64` or `amd64` (universal zip).

## Errors

- None.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Version", resp.Version, "3.6.11")
	if !strings.Contains(resp.URL, "iTerm2-3_6_11.zip") {
		t.Fatalf("URL %q missing zip name", resp.URL)
	}
	assertNoArchInURL(t, resp.URL)
}
```
