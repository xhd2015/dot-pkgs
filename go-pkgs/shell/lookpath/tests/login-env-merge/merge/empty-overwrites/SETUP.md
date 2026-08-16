# Scenario

**Feature**: empty value in a later slice overwrites a prior value

```
MergeEnvs([FOO=1], [FOO=]) -> [FOO=]
```

## Steps

1. Set MergeInputs so later FOO has empty value.

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
		{"FOO="},
	}
	return nil
}
```
