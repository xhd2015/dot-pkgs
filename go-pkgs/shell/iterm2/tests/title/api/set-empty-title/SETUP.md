# Scenario

**Feature**: SetTitle rejects empty title

```
# empty title with session env present
SetTitle("", session) -> error
```

## Steps

1. Phase set-title; session env set (not cleared); Title empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "set-title"
	req.ClearSessionEnv = false
	req.Title = ""
	req.Target = "session"
	return nil
}
```
