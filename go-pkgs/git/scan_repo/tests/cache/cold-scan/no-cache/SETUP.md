# Scenario

**Feature**: cold Scan with `NoCache=true` skips all cache writes

```
NoCache=true + CacheRoot set
  -> full live discovery
  -> no home/repos.json, no walk.jsonl, no mirror/
```

## Preconditions

- Leaves set NoCache=true.

## Steps

1. Force NoCache=true at grouping level.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoCache = true
	return nil
}
```
