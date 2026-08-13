# Scenario

**Feature**: Always enables color even when stdout is not a TTY

```
# Always ignores TTY
Resolve(Always, tty=false, noColorEnv="") -> true
```

## Steps

1. Set Mode to `Always`, TTY false, empty `noColorEnv`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = color.Always
	req.TTY = false
	req.NoColorEnv = ""
	return nil
}
```
