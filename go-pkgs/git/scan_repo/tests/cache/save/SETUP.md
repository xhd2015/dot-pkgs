# Scenario

**Feature**: `SaveCacheEntry` writes `entry.json` under the mirror tree (atomic)

```
# save path
CacheRoot + realPath + CacheEntry -> SaveCacheEntry
  -> mkdir mirror dirs, atomic write entry.json
```

## Preconditions

- Leaves set `CacheOp` to `save-load` or `overwrite`.
- Real paths use multi-segment abs forms so nested mirror dirs are exercised.

## Steps

1. Default `RealPath` to a nested absolute path for intermediate-dir coverage.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RealPath = "/Users/xhd2015/Projects/org/saved-repo"
	return nil
}
```
