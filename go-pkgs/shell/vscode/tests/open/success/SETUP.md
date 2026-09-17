# Scenario

**Feature**: a directory opens in the resolved CLI with window reuse

```
Resolve(home) -> code -> Run(["code", "-r", dir]) -> Result
```

## Steps

1. Set `Kind=dir` and `Via="candidate"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "dir"
	req.Via = "candidate"
	return nil
}
```
