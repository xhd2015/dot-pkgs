# Scenario

**Feature**: `maxDepth <= 0` defaults to 16 so a deep fixture chain is fully found

```
linearChain 1..5, root=1, maxDepth=0 -> Descendants -> all 5 nodes (default 16)
```

## Steps

1. Set `req.Op` to `"descendants"`.
2. Use `linearChainFixture()`; `RootPID=1`; `MaxDepth=0` (trigger default).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "descendants"
	req.RootPID = 1
	req.MaxDepth = 0
	req.FixtureProcs = linearChainFixture()
	return nil
}
```
