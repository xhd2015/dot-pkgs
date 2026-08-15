# Scenario

**Feature**: SetTitle / GetTitle API validation before osascript

```
# env / input validation
SetTitle / GetTitle without ITERM_SESSION_ID or with empty title -> error
```

## Steps

1. API leaves set Phase and ClearSessionEnv / Title as needed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Leaves set Phase (set-title | get-title) and error conditions.
	if req.Target == "" {
		req.Target = "session"
	}
	return nil
}
```
