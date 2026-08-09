# Scenario

**Feature**: set same KEY twice → last value wins (S3)

```
# S3 last-wins
Base has FOO=from-base
Set = FOO=first, FOO=second
  -> MergeProcessEnv
  -> FOO=second
```

## Steps

1. Base includes `FOO=from-base` and `PATH=/bin`.
2. Set lists `FOO=first` then `FOO=second`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Base = []string{
		"FOO=from-base",
		"PATH=/bin",
	}
	req.Set = []string{
		"FOO=first",
		"FOO=second",
	}
	req.Unset = nil
	return nil
}
```
