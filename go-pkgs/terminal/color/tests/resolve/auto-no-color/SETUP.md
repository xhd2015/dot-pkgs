# Scenario

**Feature**: Auto treats any non-empty noColorEnv as disable, even on a TTY

```
# Auto + TTY + NO_COLOR=1
Resolve(Auto, tty=true, noColorEnv="1") -> false
```

## Steps

1. Set Mode to `Auto`, TTY true, `noColorEnv` `"1"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = color.Auto
	req.TTY = true
	req.NoColorEnv = "1"
	return nil
}
```
