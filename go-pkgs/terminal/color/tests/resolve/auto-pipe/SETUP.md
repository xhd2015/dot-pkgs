# Scenario

**Feature**: Auto disables color when stdout is not a TTY

```
# Auto + pipe
Resolve(Auto, tty=false, noColorEnv="") -> false
```

## Steps

1. Set Mode to `Auto`, TTY false, empty `noColorEnv`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = color.Auto
	req.TTY = false
	req.NoColorEnv = ""
	return nil
}
```
