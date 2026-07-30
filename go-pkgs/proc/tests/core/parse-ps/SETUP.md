# Scenario

**Feature**: pure `ParsePSOutput` turns `ps` text into `[]Proc`

```
ps fixture bytes -> ParsePSOutput -> []Proc (PID, PPID, Cmd)
# invalid lines skipped; never panic
```

## Preconditions

- Leaves supply `req.PSOutput` bytes (inline or `testdata/`).
- No live `ps` invocation.

## Steps

1. Set `req.Op` to `"parse-ps"`.
2. Leaf fills `req.PSOutput`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "parse-ps"
	return nil
}
```
