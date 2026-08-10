# Scenario

**Feature**: base has NO_COLOR; unset NO_COLOR → absent (S8)

```
# S8 NO_COLOR
Base has NO_COLOR=1 and PATH=/bin
Unset = NO_COLOR
  -> MergeProcessEnv
  -> NO_COLOR absent; PATH preserved
```

## Steps

1. Base: `NO_COLOR=1`, `PATH=/bin`.
2. Unset: `NO_COLOR`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Base = []string{
		"NO_COLOR=1",
		"PATH=/bin",
	}
	req.Set = nil
	req.Unset = []string{"NO_COLOR"}
	return nil
}
```
