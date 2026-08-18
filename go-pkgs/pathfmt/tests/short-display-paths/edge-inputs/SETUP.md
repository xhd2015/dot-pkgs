# Scenario

**Feature**: edge inputs for normalization and fallback behavior

```
# formatter pipeline
caller path string -> Short -> Abs normalize -> cwd/home rules -> display string

# fallback
otherwise -> absolute unchanged
```

## Preconditions

- Leaves configure edge-case paths and `req.BaseDir`. No process `chdir`.

## Steps

1. Leaves set Path, BaseDir, and Op.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}
```
