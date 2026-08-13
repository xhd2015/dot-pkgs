# Scenario

**Feature**: Auto follows TTY when noColorEnv is empty

```
# Auto + TTY, no NO_COLOR
Resolve(Auto, tty=true, noColorEnv="") -> true
```

## Steps

1. Set Mode to `Auto`, TTY true, leave `noColorEnv` empty (zero value).

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
	return nil
}
```
