# Scenario

**Feature**: empty noColorEnv does not disable Auto on a TTY

```
# empty string is not NO_COLOR
Resolve(Auto, tty=true, noColorEnv="") -> true
```

## Steps

1. Set Mode to `Auto`, TTY true, explicit empty `noColorEnv` `""`.

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
	req.NoColorEnv = ""
	return nil
}
```
