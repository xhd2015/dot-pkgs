# Scenario

**Feature**: `DefaultDirs(home)` pure directory list

```
DefaultDirs(home) -> ordered search dirs (home-relative when home set + system)
```

## Steps

1. Set `Operation=default-dirs`.
2. Leaves set `DefaultDirsHome` and assert list membership/order constraints.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "default-dirs"
	return nil
}
```
