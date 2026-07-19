## Expected

- `ClassifyBusyFromComm` returns `BusyStateIdle` for common login shells:
  `zsh`, `bash`, `fish`, `sh`.
- Basename of a path (e.g. `/bin/zsh`) is also idle when the probe succeeds.

## Exit Code

- N/A (classify-busy phase)

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	shells := []string{
		"zsh", "bash", "fish", "sh",
		"/bin/zsh", "/usr/local/bin/bash",
		// login shells (ps COMM often has a leading '-')
		"-bash", "-zsh", "-fish", "-sh",
	}
	for _, name := range shells {
		got := iterm2.ClassifyBusyFromComm(name, true)
		if got != iterm2.BusyStateIdle {
			t.Fatalf("ClassifyBusyFromComm(%q, true) = %v, want BusyStateIdle", name, got)
		}
	}
	// Setup fixture path also idle
	got := iterm2.ClassifyBusyFromComm(req.FgComm, req.FgOK)
	if got != iterm2.BusyStateIdle {
		t.Fatalf("ClassifyBusyFromComm(%q, %v) = %v, want BusyStateIdle", req.FgComm, req.FgOK, got)
	}
}
```
