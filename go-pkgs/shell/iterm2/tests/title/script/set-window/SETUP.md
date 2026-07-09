# Scenario

**Feature**: BuildSetTitleScript for window target

```
BuildSetTitleScript(sessionID, window, "Hello Window")
  -> AppleScript locates UUID, sets containing window name
```

## Steps

1. Phase build-set-title; target window; title Hello Window.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "build-set-title"
	req.Target = "window"
	req.Title = "Hello Window"
	return nil
}
```
