# Scenario

**Feature**: NoCache Scan does not write walk log artifacts

```
# no-cache mode
NoCache=true + CacheRoot set
  -> Scan full walk
  -> ExpectWalkLog=false
  -> home/walk.jsonl and walk.cursor.json must not appear
```

## Preconditions

- Grouping node for NoCache leaves only.

## Steps

1. Force `NoCache=true` and `ExpectWalkLog=false`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.NoCache = true
	req.ExpectWalkLog = false
	return nil
}
```
