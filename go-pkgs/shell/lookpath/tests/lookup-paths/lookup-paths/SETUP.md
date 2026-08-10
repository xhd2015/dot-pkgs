# Scenario

**Feature**: `LookupPaths(names, opts)` — batch multi-stage resolution

```
LookupPaths(names, Options{injectables})
  -> one LookupItem per name (order preserved)
  -> Missing best-effort; empty name → error
```

## Steps

1. Set `Operation=lookup-paths`.
2. Child groups/leaves set Names and stage fixtures; assert Items + From.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "lookup-paths"
	return nil
}
```
