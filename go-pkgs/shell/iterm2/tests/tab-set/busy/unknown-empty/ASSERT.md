## Expected

- Empty `fgComm` → `BusyStateUnknown` (regardless of ok).
- `ok == false` → `BusyStateUnknown` even if comm looks like a shell.
- Whitespace-only comm → `BusyStateUnknown`.

## Exit Code

- N/A (classify-busy phase)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {

	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		comm string
		ok   bool
	}{
		{"", true},
		{"", false},
		{"zsh", false},
		{"node", false},
		{"   ", true},
	}
	for _, c := range cases {
		got := iterm2.ClassifyBusyFromComm(c.comm, c.ok)
		if got != iterm2.BusyStateUnknown {
			t.Fatalf("ClassifyBusyFromComm(%q, %v) = %v, want BusyStateUnknown", c.comm, c.ok, got)
		}
	}
	got := iterm2.ClassifyBusyFromComm(req.FgComm, req.FgOK)
	if got != iterm2.BusyStateUnknown {
		t.Fatalf("fixture ClassifyBusyFromComm(%q, %v) = %v, want BusyStateUnknown", req.FgComm, req.FgOK, got)
	}
}
```
