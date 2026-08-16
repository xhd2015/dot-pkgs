# Scenario

**Feature**: later slice overwrites the same key

```
MergeEnvs([FOO=1], [FOO=2]) -> [FOO=2]
```

## Steps

1. Set MergeInputs to two slices with FOO.

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
		{"FOO=2"},
	}
	return nil
}
```
