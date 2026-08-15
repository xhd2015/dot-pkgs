# Scenario

**Feature**: BuildSetTitleScript for session target

```
BuildSetTitleScript(sessionID, session, "Hello Tab")
  -> AppleScript locates UUID, sets session name to Hello Tab
```

## Steps

1. Phase build-set-title; target session; title Hello Tab.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-set-title"
	req.Target = "session"
	req.Title = "Hello Tab"
	return nil
}
```
