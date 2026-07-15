# Scenario

**Feature**: DiffOwnedTrees detects blob identity changes between snapshots

```
OwnedTree old vs new -> DiffOwnedTrees -> Changed bool
```

## Steps

1. Set `req.Op` to `"diff"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "diff"
	return nil
}
```