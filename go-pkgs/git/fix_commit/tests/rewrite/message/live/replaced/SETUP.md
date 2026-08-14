# Scenario

**Feature**: `-m` replaces the message; author, dates, tree, and parent stay

```
RunCLI <sha> -m "corrected message" -> new commit -> master tip moves
```

## Steps

1. Append `-m` `corrected message` to `req.Args`.
2. Expect new SHA, subject `corrected message`, author still Alice, `master`
   at the new SHA.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "-m", "corrected message")
	return nil
}
```
