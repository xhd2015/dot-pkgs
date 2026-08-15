# Scenario

**Feature**: Scan path applies liveness so dead indexed repos never emit

```
# cold indexes live + doomed; remove doomed/.git
cold seed both
  -> remove doomed/.git (or strip git marker)
  -> Scan warm
  -> Result has live only; doomed omitted
```

## Preconditions

- Distinct from pure `ApplyLiveness` leaf under `cache/repo-index/liveness/` —
  this exercises **Scan** end-to-end with index/mirror candidates.

## Steps

1. Leaves cold-seed two repos, kill one `.git`, set `DeadPath`.

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
