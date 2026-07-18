# Scenario

**Feature**: warm Scan serves live repos from the durable index

```
# Setup: cold seed (mirror + index) then Run second Scan
cold Scan seeds home/repos.json
  -> second Scan (warm-eligible, NoCache=false, Refresh=false)
  -> Result includes indexed live mains
  -> index still loadable
```

## Preconditions

- Leaves call `coldSeedScan` then leave FS unchanged (or only non-repo changes).
- Run is the **second** Scan under test.

## Steps

1. Default warm-eligible options.
2. Leaves plant workspace, cold-seed, set Roots.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.NoCache = false
	req.Refresh = false
	return nil
}
```
