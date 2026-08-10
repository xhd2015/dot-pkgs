# Scenario

**Feature**: BuildFocusScript activates iTerm and selects window + tab

```
SessionRef{WindowID: win-focus-42, TabIndex: 2}
  -> BuildFocusScript
  -> activate; select window by id; select tab 2; no create window
```

## Steps

1. Phase `build-focus-script`.
2. FocusRef from parent grouping.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-focus-script"
	return nil
}
```
