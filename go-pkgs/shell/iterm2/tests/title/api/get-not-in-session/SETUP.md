# Scenario

**Feature**: GetTitle errors when not inside iTerm2

```
# ITERM_SESSION_ID unset
GetTitle(session) -> error
```

## Steps

1. Phase get-title; clear session env.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "get-title"
	req.ClearSessionEnv = true
	req.Target = "session"
	return nil
}
```
