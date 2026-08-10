# Scenario

**Feature**: `ApplyDirs` with zero dirs is a success no-op

```
empty target + ApplyDirs(target) -> success; target stays empty
```

## Steps

1. Leave dirs list empty (no LayerSpecs).
2. Run ApplyDirs with no directory arguments.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Layers = nil
	req.DirsRel = nil
	return nil
}
```
