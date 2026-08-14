# Scenario

**Feature**: `--help` prints usage including every flag

```
RunCLI --help -> locked usage on stdout -> exit 0
```

## Steps

1. Set `req.Args` to `["--help"]`.
2. Expect the locked usage text (DSN), trailing newline, empty stderr.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"--help"}
	return nil
}
```
