# Scenario

**Feature**: AppleScript string escaping helpers

```
escape input -> EscapePathForAppleScript / EscapeCommandForAppleScript -> escaped string
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Phase == "" {
		req.Phase = "escape-path"
	}
	return nil
}
```