# Scenario

**Feature**: Never disables color even when stdout is a TTY

```
# Never ignores TTY
Resolve(Never, tty=true, noColorEnv="") -> false
```

## Steps

1. Set Mode to `Never`, TTY true, empty `noColorEnv`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = color.Never
	req.TTY = true
	req.NoColorEnv = ""
	return nil
}
```
