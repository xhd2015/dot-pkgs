# Scenario

**Feature**: unknown flag is a fatal parse error

```
RunCLI HEAD -m x --nope -> Error: unknown flag: --nope
```

## Steps

1. Set `req.Args` to `["HEAD", "-m", "x", "--nope"]`.
2. Flag parse fails before repo access.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"HEAD", "-m", "x", "--nope"}
	return nil
}
```
