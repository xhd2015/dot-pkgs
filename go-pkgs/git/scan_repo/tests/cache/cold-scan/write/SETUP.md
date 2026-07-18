# Scenario

**Feature**: cold Scan with `NoCache=false` seeds durable index (and walk log), not mirror

```
# write-enabled cold path
NoCache=false + CacheRoot set
  -> Scan full walk
  -> home/repos.json lists discovered checkouts
  -> no mirror/ tree
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
