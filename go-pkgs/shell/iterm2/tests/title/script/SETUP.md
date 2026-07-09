# Scenario

**Feature**: BuildSetTitleScript / BuildGetTitleScript emit AppleScript

```
# build scripts (no osascript)
Build*TitleScript(sessionID, target, …) -> AppleScript string
```

## Steps

1. Ensure a default session id for script leaves; leaves set Phase and target.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = defaultTitleSessionID
	}
	return nil
}
```
