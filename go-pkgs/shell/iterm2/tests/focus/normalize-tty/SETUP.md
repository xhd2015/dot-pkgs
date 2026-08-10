# Scenario

**Feature**: NormalizeTTY equates bare and /dev TTY forms

```
NormalizeTTY("ttys148") == NormalizeTTY("/dev/ttys148")
NormalizeTTY("") == ""
```

## Steps

1. Leaves set Phase `normalize-tty`.
2. Forms leaf asserts equality table in Assert (may also use Run for one input).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "normalize-tty"
	return nil
}
```
