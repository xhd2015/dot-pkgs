# Scenario

**Feature**: cold Scan with `NoCache=false` writes mirror entries for visited dirs

```
# write-enabled cold path
NoCache=false + CacheRoot set
  -> Scan full walk
  -> mirror entry.json for repos and intermediate directories
```

## Preconditions

- `req.NoCache` is false.
- Leaves build workspace under `t.TempDir()` and set `req.Roots`.

## Steps

1. Force `NoCache=false` (cache write enabled).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.NoCache = false
	return nil
}
```
