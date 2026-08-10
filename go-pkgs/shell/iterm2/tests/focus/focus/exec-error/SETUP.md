# Scenario

**Feature**: Focus propagates Exec errors

```
Focus(ref, FocusConfig{Exec: returns err})
  -> same error (non-nil)
```

## Steps

1. Phase `focus`.
2. Mock Exec returns a sentinel error; Assert checks propagation.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "focus"
	req.ExecError = "mock osascript failed"
	return nil
}
```
