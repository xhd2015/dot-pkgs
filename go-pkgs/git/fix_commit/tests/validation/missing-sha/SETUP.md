# Scenario

**Feature**: SHA is a required positional

```
RunCLI -m x -> Error: fix-commit requires <sha>
```

## Steps

1. Set `req.Args` to `["-m", "x"]` (change flag present, SHA absent).
2. Expect fatal before any git work.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"-m", "x"}
	return nil
}
```
