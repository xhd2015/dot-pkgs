# Scenario

**Feature**: `-m` equal to the current message is a no-op error

```
RunCLI <sha> -m "fix typo" -> Error: nothing to change
```

## Steps

1. Append `-m` with the existing subject `fix typo`.
2. HEAD must stay at the old SHA.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "-m", "fix typo")
	return nil
}
```
