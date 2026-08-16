# Scenario

**Feature**: nil middle slice is skipped (same as MergeEnvs(a, b))

```
MergeEnvs([FOO=1], nil, [BAR=2]) -> [FOO=1, BAR=2]
```

## Steps

1. MergeInputs with a nil middle slice.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.MergeInputs = [][]string{
		{"FOO=1"},
		nil,
		{"BAR=2"},
	}
	return nil
}
```
