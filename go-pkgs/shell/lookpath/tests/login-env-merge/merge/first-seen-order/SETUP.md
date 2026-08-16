# Scenario

**Feature**: overwrite keeps first-seen key order (does not move key to end)

```
MergeEnvs([A=1, B=1], [A=2]) -> [A=2, B=1]
```

## Steps

1. First slice introduces A then B; later overwrites A only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.MergeInputs = [][]string{
		{"A=1", "B=1"},
		{"A=2"},
	}
	return nil
}
```
