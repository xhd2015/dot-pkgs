# Scenario

**Feature**: bare ttysN and /dev/ttysN normalize equal; empty stays empty

```
NormalizeTTY("ttys148") <-> NormalizeTTY("/dev/ttys148")  # equal, non-empty
NormalizeTTY("") -> ""
NormalizeTTY("   ") -> "" or equal only to its own trim policy (empty-ish)
```

## Steps

1. Phase already `normalize-tty` from parent.
2. Seed one Run input (`/dev/ttys148`) so Response.Normalized is exercised;
   Assert covers the full table via direct `NormalizeTTY` calls.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TTY = "/dev/ttys148"
	return nil
}
```
