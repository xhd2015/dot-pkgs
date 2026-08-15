# Scenario

**Feature**: title strings use AppleScript literal escaping

```
# escape title for embedded string literals
title with quotes/backslashes -> EscapePathForAppleScript (or title escaper)
```

## Steps

1. Phase escape-title for descendants.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "escape-title"
	return nil
}
```
