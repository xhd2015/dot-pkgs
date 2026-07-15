# Scenario

**Feature**: `MirrorEntryPath` maps a real directory path to the on-disk entry file

```
# path mapping only — no read/write of entry content
realPath + CacheRoot -> MirrorEntryPath -> .../mirror/<rel>/entry.json
```

## Preconditions

- `req.CacheOp` is `"mirror-path"`.
- Real path need not exist on disk for pure mapping.

## Steps

1. Set `CacheOp` to `mirror-path`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = "mirror-path"
	return nil
}
```
