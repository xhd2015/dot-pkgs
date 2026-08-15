# Scenario

**Feature**: full successful cold Scan produces walk log under home/

```
# cold mode
NoCache=false + CacheRoot set
  -> Scan cold
  -> ExpectWalkLog=true
  -> home/walk.jsonl and walk.cursor.json must exist after Run
```

## Preconditions

- Grouping node for cache-enabled cold walk leaves only.
- Does not plant fixtures; leaves create workspace roots.

## Steps

1. Force `NoCache=false` and `ExpectWalkLog=true`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoCache = false
	req.ExpectWalkLog = true
	return nil
}
```
