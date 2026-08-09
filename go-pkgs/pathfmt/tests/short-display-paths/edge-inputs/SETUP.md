# Scenario

**Feature**: edge inputs for normalization and fallback behavior

```
# formatter pipeline
caller path string -> Short -> Abs normalize -> cwd/home rules -> display string

# fallback
otherwise -> absolute unchanged
```

## Preconditions

- Cwd is saved and restored; leaves configure edge-case paths.

## Steps

1. Save and restore the original cwd.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	saveAndRestoreCwd(t)
	return nil
}```
