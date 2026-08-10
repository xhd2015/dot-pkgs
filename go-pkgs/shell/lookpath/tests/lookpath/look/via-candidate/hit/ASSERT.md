## Expected

- `err == nil`.
- `resp.Path` equals the candidate absolute path.
- `resp.Via == "candidate"`.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertPathEqual(t, resp.Path, req.ExtraCandidates[0])
	assertEqual(t, "Via", resp.Via, "candidate")
}
```
