# Scenario

**Feature**: SaveRepoIndex then LoadRepoIndex preserves schema v1 fields

```
# round-trip
SaveRepoIndex(index) -> LoadRepoIndex(cacheRoot, universe)
  -> same version, universe, base, updated_at, repos[] fields
  -> file at <CacheRoot>/<universe>/repos.json
```

## Preconditions

- Leaves set `IndexOp` to `save-load` and a concrete `Universe`.
- Index documents use non-default values so zero-value bugs fail loudly.

## Steps

1. Set `IndexOp` to `save-load`.
2. Child leaves fill `Universe` and `Index`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.IndexOp = "save-load"
	return nil
}
```
