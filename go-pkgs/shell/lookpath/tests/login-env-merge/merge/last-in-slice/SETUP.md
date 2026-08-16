# Scenario

**Feature**: within one slice, last occurrence of a key wins

```
MergeEnvs([FOO=1, FOO=2]) -> [FOO=2]
```

## Steps

1. Single slice with FOO twice.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.MergeInputs = [][]string{
		{"FOO=1", "FOO=2"},
	}
	return nil
}
```
