# Scenario

**Feature**: AppleScript string escaping helpers

```
escape input -> EscapePathForAppleScript / EscapeCommandForAppleScript -> escaped string
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Phase == "" {
		req.Phase = "escape-path"
	}
	return nil
}
```