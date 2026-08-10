# Scenario

**Feature**: unset then set KEY → KEY present with set value (S5)

```
# S5 reintroduce
Base has FOO=old, PATH=/bin
Unset = FOO
Set = FOO=new
  -> MergeProcessEnv
  -> FOO=new (not old; not absent)
```

## Steps

1. Base: `FOO=old`, `PATH=/bin`.
2. Unset: `FOO`.
3. Set: `FOO=new`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Base = []string{
		"FOO=old",
		"PATH=/bin",
	}
	req.Unset = []string{"FOO"}
	req.Set = []string{"FOO=new"}
	return nil
}
```
