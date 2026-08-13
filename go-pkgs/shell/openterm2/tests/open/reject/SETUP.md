# Scenario

**Feature**: invalid `dir` errors before either opener runs

```
empty | whitespace | file | missing
  -> OpenConfig -> error
OpenITerm and OpenTerminal never called
```

## Steps

1. Arm `ResolveITerm` with a fake iTerm.app path so a skipped validation
   would call `OpenITerm` (Assert requires neither opener).
2. Leaves replace `Dir` with an invalid value / fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermAppPath = fakeITermApp(req)
	return nil
}
```
