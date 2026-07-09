# Scenario

**Feature**: title strings use AppleScript literal escaping

```
# escape title for embedded string literals
title with quotes/backslashes -> EscapePathForAppleScript (or title escaper)
```

## Steps

1. Phase escape-title for descendants.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "escape-title"
	return nil
}
```
