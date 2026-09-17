# Scenario

**Feature**: `*Config` builds argv, then calls the injected runner once

```
PathConfig / URLConfig / AppConfig -> build argv -> Run(argv) -> Result
```

## Steps

1. Set `Operation=run`.
2. Leaves set `Kind`, the target fields, and optionally `RunnerErr`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "run"
	return nil
}
```
