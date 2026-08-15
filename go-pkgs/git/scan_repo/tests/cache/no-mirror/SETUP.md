# Scenario

**Feature**: dense `mirror/` cache is retired — Scan must not create or grow it

```
# cold + warm with CacheRoot (v2: index + walk log only)
caller roots + CacheRoot + NoCache=false -> Scan
  -> home/repos.json + home/walk.jsonl (allowed)
  -> <CacheRoot>/mirror MUST NOT exist (or must stay empty / absent)
```

## Preconditions

- `CacheOp` empty; Scan path only.
- Explicit temp `CacheRoot` from parent `cache/SETUP.md`.
- Leaves prove absence of dense mirror after cold and after warm second Scan.

## Steps

1. Leave CacheOp empty; NoCache=false by default for write-enabled Scans.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CacheOp = ""
	req.NoCache = false
	req.ListRemotes = false
	req.ListWorktrees = false
	return nil
}
```
