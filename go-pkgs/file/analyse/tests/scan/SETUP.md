# Scenario

**Feature**: `analyse.Scan` discovers and measures fake HOME children

```
# direct scan — no HTTP/SSE
caller Options{Home} -> Scan -> []EntryResult + ScanSummary
```

## Preconditions

- `Mode` remains `scan` (default) for every leaf in this branch.
- Each leaf seeds its own temp HOME via `SeedProfile`.

## Steps

1. Force `req.Mode = "scan"`.
2. Leaves set `SeedProfile` and optional OnEntry fields.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "scan"
	return nil
}
```