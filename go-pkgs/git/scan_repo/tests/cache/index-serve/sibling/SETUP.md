# Scenario

**Feature**: warm Scan discovers new sibling checkouts of indexed repos

```
# after cold has parent/A, plant parent/B/.git
cold: parent/A only in index/mirror
plant parent/B with .git (never cold-written)
  -> Scan warm (Refresh=false)
  -> Result includes A and B  # sibling ReadDir, not full re-cold only
```

## Preconditions

- Contrast: classic mirror warm (`cache/warm/serves-cached-omits-new`) omits
  brand-new; P2 sibling probe **includes** a sibling of an indexed repo.
- `Refresh` stays false so this is not force-cold.

## Steps

1. Leaves cold-seed with only repo A, plant sibling B, set paths on Request.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.NoCache = false
	req.Refresh = false
	return nil
}
```
