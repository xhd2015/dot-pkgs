# Scenario

**Feature**: disjoint keys are unioned; later keys append after first-seen

```
MergeEnvs([FOO=1], [BAR=2]) -> [FOO=1, BAR=2]
```

## Steps

1. Set MergeInputs to FOO then BAR slices.

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
		{"BAR=2"},
	}
	return nil
}
```
