# Scenario

Root setup for shell/applescript doctests. Leaves set `Op` and inputs.

```
# pure
EscapeString / CheckWriteText / DocumentWriteTextLimitation

# live (label e2e)
iterm2.OpenConfig ForceNew + FollowUp → file compare
```

## Preconditions

- Module `github.com/xhd2015/dot-pkgs/go-pkgs`.
- Live leaves require darwin + iTerm2 installed.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if req.Op == "" {
		req.Op = "check"
	}
	return nil
}
```
