# Scenario

**Feature**: BuildGetTitleScript for window target

```
BuildGetTitleScript(sessionID, window)
  -> AppleScript returns window name for session's window
```

## Steps

1. Phase build-get-title; target window.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "build-get-title"
	req.Target = "window"
	return nil
}
```
