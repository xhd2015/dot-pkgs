# Scenario

**Feature**: unset key not in base is a no-op (base otherwise intact)

```
# absent key unset
Base has PATH=/bin, HOME=/tmp/home (no MISSING)
Unset = MISSING
  -> MergeProcessEnv
  -> PATH and HOME unchanged; MISSING still absent
```

## Steps

1. Base without `MISSING`.
2. Unset lists `MISSING` only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Base = []string{
		"PATH=/bin",
		"HOME=/tmp/home",
	}
	req.Set = nil
	req.Unset = []string{"MISSING"}
	return nil
}
```
