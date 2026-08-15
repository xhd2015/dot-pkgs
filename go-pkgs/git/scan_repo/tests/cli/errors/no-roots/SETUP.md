# Scenario

**Feature**: missing required `--root` flag

```
RunCLI (no --root) -> validation error on stderr
```

## Steps

1. Invoke with empty argv (no `--root`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{}
	return nil
}
```