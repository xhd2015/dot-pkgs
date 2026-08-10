# Scenario

**Feature**: Focus runs focus script via injectable Exec

```
Focus(ref, FocusConfig{Exec: capture})
  -> Exec called with script containing window id and tab index
  -> nil error
```

## Steps

1. Phase `focus` (product invoked in Assert).
2. FocusRef from parent; mock Exec records scripts.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "focus"
	req.ExecError = ""
	return nil
}
```
