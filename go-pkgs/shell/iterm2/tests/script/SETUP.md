# Scenario

**Feature**: `BuildScript` emits smart-open AppleScript

```
caller dir + follow-ups -> BuildScript -> AppleScript (structure + write text lines)
```

## Steps

1. Set `req.Phase` to `build-script`.
2. Set `req.Dir` to an absolute fixture path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "build-script"
	return nil
}
```