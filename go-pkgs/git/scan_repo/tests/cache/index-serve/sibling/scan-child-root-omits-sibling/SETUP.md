# Scenario

**Bug**: warm Scan with `Roots: [A]` must not emit neighboring sibling checkout `B`

```
# parent holds two independent mains; scan root is only child A
parent/A/.git  --cold seed--> indexed (warm-eligible)
parent/B/.git  present (sibling of A; outside absRoot=A)
  -> Scan warm Roots=[abs(A)] (Refresh=false)
  -> sibling ReadDir(parent) *would* see B
  -> Result paths ⊆ under A; B omitted  # pathIsUnderRoot filter after probe
```

## Preconditions

- Contrast: `discovers-new` scans the **parent** and expects A+B.
  This leaf scans **child A only** and requires B never appear.
- Sibling probe still runs (same parent `ReadDir`); the contract is post-probe
  filtering with `pathIsUnderRoot(absRoot, path)`.
- `Refresh` stays false so this is warm index-serve + sibling, not force-cold.
- Classic TDD: current code merges sibling discoveries without under-root drop → RED.

## Steps

1. Create `parent/A` and `parent/B` with fake `.git`.
2. Cold-seed with `Roots: [A]` so A is in home index (warm-eligible).
3. Set `Roots: [A]` for the warm Run; stash `KnownPath=A`, `SiblingPath=B`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	parent := t.TempDir()
	a := filepath.Join(parent, "A")
	b := filepath.Join(parent, "B")
	mkdirAll(t, a)
	mkdirAll(t, b)
	fakeGitRepo(t, a)
	fakeGitRepo(t, b)

	aAbs := absPath(t, a)
	bAbs := absPath(t, b)

	// Cold-seed against child A only: index contains A under Base=A.
	// Parent ReadDir still lists B next to A for the warm sibling probe.
	req.Roots = []string{aAbs}
	req.NoCache = false
	req.Refresh = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	req.KnownPath = aAbs
	req.SiblingPath = bAbs
	return nil
}
```
