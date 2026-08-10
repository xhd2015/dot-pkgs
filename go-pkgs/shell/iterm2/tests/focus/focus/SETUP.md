# Scenario

**Feature**: BuildFocusScript + Focus with injectable Exec

```
SessionRef
  -> BuildFocusScript -> activate + select window + select tab
  -> Focus(cfg.Exec) -> Exec(script); error propagates
```

## Steps

1. Leaves set Phase (`build-focus-script` or `focus`) and `FocusRef`.
2. Focus leaves call product `Focus` in Assert with mock Exec.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Shared focus target for build-script / inject-exec / exec-error leaves.
	if req.FocusRef.WindowID == "" {
		req.FocusRef = SessionRefInput{
			WindowID:   "win-focus-42",
			WindowName: "FocusWin",
			TabIndex:   2,
			SessionID:  "sess-focus",
			TTY:        "/dev/ttys148",
			Name:       "Target",
		}
	}
	return nil
}
```
