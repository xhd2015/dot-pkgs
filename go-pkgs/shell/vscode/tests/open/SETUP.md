# Scenario

**Feature**: `OpenConfig` / `OpenFileConfig` resolve once, run once

```
*Config -> Resolve(home) -> args -> Run(args) -> Result
```

## Steps

1. Set `Operation=open`.
2. Leaves set `Kind=dir|file`, window mode / line, and failure injection.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "open"
	return nil
}
```
