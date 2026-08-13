# Scenario

**Feature**: Always wins over a non-empty noColorEnv (flags win)

```
# flags win — NO_COLOR does not disable Always
Resolve(Always, tty=true, noColorEnv="1") -> true
```

## Steps

1. Set Mode to `Always`, TTY true, `noColorEnv` `"1"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = color.Always
	req.TTY = true
	req.NoColorEnv = "1"
	return nil
}
```
