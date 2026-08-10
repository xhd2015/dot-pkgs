# Scenario

**Feature**: BuildSessionListScript scans windows/tabs/sessions and emits tty

```
BuildSessionListScript()
  -> AppleScript mentions windows, tabs, sessions, tty
  -> uses ASCII TAB field separator (not bare tab inside iTerm tell)
```

## Steps

1. Phase `build-session-list-script`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-session-list-script"
	return nil
}
```
