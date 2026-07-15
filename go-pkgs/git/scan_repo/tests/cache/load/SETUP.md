# Scenario

**Feature**: `LoadCacheEntry` reads mirror `entry.json` or reports missing/corrupt

```
# load-only
CacheRoot + realPath -> LoadCacheEntry -> (entry, ok, err)
```

## Preconditions

- `req.CacheOp` is `"load"`.
- Leaves arrange on-disk state under the mirror path (missing, corrupt, or none).

## Steps

1. Set `CacheOp` to `load`.
2. Default `RealPath` to a stable absolute path under a synthetic home-style root
   (leaves may override).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = "load"
	req.RealPath = "/Users/xhd2015/Projects/load-target"
	return nil
}
```
