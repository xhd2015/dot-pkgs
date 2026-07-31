# Scenario

**Feature**: list all iTerm sessions via script + pure parse

```
BuildSessionListScript -> AppleScript dump (windows/tabs/sessions + tty)
stdout -> ParseSessionListOutput -> []SessionRef
```

## Preconditions

- Dump format: `WindowID\tWindowName\tTabIndex\tSessionIndex\tSessionID\tTTY\tName`
- Field separator in script: ASCII character 9 (not bare AppleScript `tab`).

## Steps

1. Parse leaves set Phase `parse-session-list` and `ListOutput`.
2. Build leaf sets Phase `build-session-list-script`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves set Phase.
	return nil
}
```
