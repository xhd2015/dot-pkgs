# Scenario

**Feature**: cold Scan path seeds the durable home universe index

```
# first Scan (cold) with CacheRoot
workspace mains -> Scan -> SaveRepoIndex(home)
  -> <CacheRoot>/home/repos.json exists with discovered main paths
```

## Preconditions

- No prior seed; Run is the cold Scan under test.
- `NoCache=false`, `Refresh=false`.

## Steps

1. Leave `NoCache=false` so cold write + index seed are enabled.
2. Leaves build workspace and set `Roots`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoCache = false
	req.Refresh = false
	return nil
}
```
