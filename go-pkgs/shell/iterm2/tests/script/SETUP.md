# Scenario

**Feature**: `BuildScript` emits smart-open AppleScript

```
caller dir + follow-ups -> BuildScript -> AppleScript (structure + write text lines)
```

## Steps

1. Set `req.Phase` to `build-script`.
2. Set `req.Dir` to an absolute fixture path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-script"
	return nil
}
```