# Scenario

**Feature**: empty lines and garbage `ps` lines are skipped without panic

```
mixed valid + junk bytes -> ParsePSOutput -> only valid PID/PPID rows
```

## Steps

1. Set `req.PSOutput` to a blob with blank lines, non-numeric fields, and one
   valid row.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// blank, short, non-numeric pid, then one valid row, trailing junk
	req.PSOutput = []byte("\n\n  not-a-pid  1 /bin/x\n42\n  7   1 /bin/true\ngarbage only\n")
	return nil
}
```
