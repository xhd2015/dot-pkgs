# Scenario

**Feature**: DiffOwnedTrees detects blob identity changes between snapshots

```
OwnedTree old vs new -> DiffOwnedTrees -> Changed bool
```

## Steps

1. Set `req.Op` to `"diff"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "diff"
	return nil
}
```