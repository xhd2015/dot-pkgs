# Scenario

**Feature**: warm Scan sibling probe under a scan root (discover vs under-root filter)

```
# discovers-new: after cold has parent/A, plant parent/B/.git; scan parent
cold: parent/A only in index
plant parent/B with .git (never cold-written)
  -> Scan warm Roots=[parent] (Refresh=false)
  -> Result includes A and B  # sibling ReadDir, not full re-cold only

# scan-child-root-omits-sibling: scan absRoot=A; B is outside root
parent/A and parent/B both present; cold seeded A
  -> Scan warm Roots=[A]
  -> sibling ReadDir may see B, but Result ⊆ under A  # pathIsUnderRoot
```

## Preconditions

- Contrast: classic mirror warm (`cache/warm/serves-cached-omits-new`) omits
  brand-new; P2 sibling probe **includes** a sibling of an indexed repo when
  that sibling lies **under** the scan root.
- When scan root is a single checkout, neighboring checkouts outside the root
  must not appear (under-root filter after sibling probe).
- `Refresh` stays false so this is not force-cold.

## Steps

1. Leaves cold-seed, plant siblings / set Roots, stash paths on Request.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.NoCache = false
	req.Refresh = false
	return nil
}
```
