# Scenario

**Feature**: `Debug=true` enables structured `scan:` phase logs on stderr

```
# debug on
caller Debug=true + CacheRoot + roots
  -> Scan
  -> stderr contains scan: phase lines (mode depends on warm eligibility)
```

## Preconditions

- Parent allocated temp `CacheRoot`.
- Leaves set workspace `Roots` and choose cold (empty cache) vs warm (seed).

## Steps

1. Set `req.Debug=true` for all descendants.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Debug = true
	return nil
}
```
