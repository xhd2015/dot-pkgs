# Scenario

**Feature**: an empty directory never reaches the resolver

```
OpenConfig("  ", cfg) -> error
Resolve and Run not called
```

## Steps

1. Set `Kind=dir` and a whitespace-only directory.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "dir"
	req.Dir = "   "
	return nil
}
```
