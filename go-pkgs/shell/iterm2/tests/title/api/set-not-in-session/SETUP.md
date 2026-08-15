# Scenario

**Feature**: SetTitle errors when not inside iTerm2

```
# ITERM_SESSION_ID unset
SetTitle("x", session) -> error (not in iTerm2 session)
```

## Steps

1. Phase set-title; clear session env; non-empty title.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "set-title"
	req.ClearSessionEnv = true
	req.Title = "x"
	req.Target = "session"
	return nil
}
```
