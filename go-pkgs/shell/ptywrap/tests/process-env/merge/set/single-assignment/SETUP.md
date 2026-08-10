# Scenario

**Feature**: set single KEY=v → KEY present with v (S2)

```
# S2 single set
Base has PATH; Set has FOO=bar
  -> MergeProcessEnv
  -> FOO=bar and base PATH preserved
```

## Steps

1. Base: `PATH=/bin`, `HOME=/tmp/home`.
2. Set: `FOO=bar`.
3. Unset empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Base = []string{
		"PATH=/bin",
		"HOME=/tmp/home",
	}
	req.Set = []string{"FOO=bar"}
	req.Unset = nil
	return nil
}
```
