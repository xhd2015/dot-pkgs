## Expected

- Four process rows in file order:
  - `{1, 0, "/sbin/launchd"}`
  - `{100, 1, "/usr/bin/bash -l"}`
  - `{200, 100, "/usr/bin/vim /tmp/note.go"}`
  - `{201, 100, "sleep 30"}`

## Errors

- `err` is nil.
- Missing rows or wrong Cmd (including dropped argv tokens) is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	want := []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "/sbin/launchd"},
		{PID: 100, PPID: 1, Cmd: "/usr/bin/bash -l"},
		{PID: 200, PPID: 100, Cmd: "/usr/bin/vim /tmp/note.go"},
		{PID: 201, PPID: 100, Cmd: "sleep 30"},
	}
	assertProcsEqual(t, resp.Procs, want)
}
```
