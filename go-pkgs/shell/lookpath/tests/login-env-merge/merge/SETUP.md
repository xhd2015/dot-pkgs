# Scenario

**Feature**: `MergeEnvs` pure last-wins merge of KEY=value slices

```
MergeEnvs(a, b, …)
  later slice wins on same key
  first-seen key order; nil slices skipped
```

## Steps

1. Set `Op=merge`.
2. Leaves set `MergeInputs` or `MergeNoArgs`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "merge"
	return nil
}
```
