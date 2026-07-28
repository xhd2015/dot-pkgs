## Expected

- Non-shell foreground names map to `BusyStateBusy` (e.g. `node`, `spl`, `python3`, `vim`).

## Exit Code

- N/A (classify-busy phase)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d

	if err != nil {
		t.Fatal(err)
	}
	busy := []string{"node", "spl", "python3", "vim", "go"}
	for _, name := range busy {
		got := iterm2.ClassifyBusyFromComm(name, true)
		if got != iterm2.BusyStateBusy {
			t.Fatalf("ClassifyBusyFromComm(%q, true) = %v, want BusyStateBusy", name, got)
		}
	}
	got := iterm2.ClassifyBusyFromComm(req.FgComm, req.FgOK)
	if got != iterm2.BusyStateBusy {
		t.Fatalf("ClassifyBusyFromComm(%q, %v) = %v, want BusyStateBusy", req.FgComm, req.FgOK, got)
	}
}
```
