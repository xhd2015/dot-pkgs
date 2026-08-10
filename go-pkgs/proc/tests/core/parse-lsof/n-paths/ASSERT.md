## Expected

- Paths (first-seen unique, absolute only):
  1. `/private/tmp/workdir`
  2. `/usr/bin/vim`
  3. `/Users/me/.grok/sessions/abc/events.jsonl`
  4. `/var/log/system.log`
- Duplicate `n/private/tmp/workdir` appears once.

## Errors

- `err` is nil.
- Reordered paths, duplicates, or inclusion of `p`/`f` descriptor lines is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/private/tmp/workdir",
		"/usr/bin/vim",
		"/Users/me/.grok/sessions/abc/events.jsonl",
		"/var/log/system.log",
	}
	assertStringsEqual(t, resp.Paths, want)
}
```
