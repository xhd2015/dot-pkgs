# Scenario

**Feature**: MergeEnvs() with zero args returns empty result

```
MergeEnvs() -> nil or empty slice
```

## Steps

1. Set MergeNoArgs=true so Run calls MergeEnvs() with no arguments.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.MergeNoArgs = true
	req.MergeInputs = nil
	return nil
}
```
