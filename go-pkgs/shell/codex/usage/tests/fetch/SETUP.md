# Scenario

**Feature**: Fetch orchestration (codex endpoint, wham fallback, both fail)

```
Fetch(GetJSON inject) -> Snapshot | error
```

## Steps

1. Set `Operation=fetch`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "fetch"
	return nil
}
```
