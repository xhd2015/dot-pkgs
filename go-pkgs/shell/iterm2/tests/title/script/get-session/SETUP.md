# Scenario

**Feature**: BuildGetTitleScript for session target

```
BuildGetTitleScript(sessionID, session)
  -> AppleScript returns session name for UUID
```

## Steps

1. Phase build-get-title; target session.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "build-get-title"
	req.Target = "session"
	return nil
}
```
