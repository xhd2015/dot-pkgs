# Scenario

**Feature**: LoadRepoIndex reads universe file without write

```
# load path
LoadRepoIndex(cacheRoot, universe) -> (index, ok, err)
```

## Preconditions

- Leaves set `IndexOp` to `load`.
- Fresh `CacheRoot` from parent (no prior Save unless a leaf writes one).

## Steps

1. Set `IndexOp` to `load`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.IndexOp = "load"
	return nil
}
```
