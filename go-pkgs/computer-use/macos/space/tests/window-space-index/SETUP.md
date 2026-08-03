# Scenario

**Feature**: map CG/iTerm window id → 0-based user Desktop index (SkyLight, first display, type-0)

```
# pure helpers (no WindowServer)
Spaces[] -> BuildUserSpaceIndex -> map[spaceID]index
map + windowSpaceIDs -> ResolveWindowSpaceIndex -> index | error

# injected L2 pipeline
windowID + ManagedDisplays + CopySpaces(mask=7)
  -> SpaceIndexForWindow -> 0-based index | detectable error
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space` is importable.
- **Classic TDD:** public symbols below are **not** implemented yet; Run references them so the tree is RED until implementer lands them:
  - `SpaceInfo`, `DisplaySpaces`, `WindowSpaceOption`
  - `WithManagedDisplays`, `WithWindowSpaceIDs`, `WithCopySpacesForWindows`, `WithPlatformGOOS`
  - `SpaceIndexForWindow`, `BuildUserSpaceIndex`, `ResolveWindowSpaceIndex`
  - `ErrNotUserSpace`, `ErrSpaceNotFound` (plus existing `ErrUnsupportedPlatform`)
- No live Accessibility, iTerm, or WindowServer required for default leaves.
- Parallel-safe: inject platform via `WithPlatformGOOS`; do not use process-global env in harness.

## Steps

1. Leaf / grouping Setups fill `req.Phase` and fixtures (spaces, window id, inject results).
2. Root `Run` dispatches on `req.Phase` to pure helpers or `SpaceIndexForWindow`.

## Context

- Canonical fixture (requirement): first-display type-0 ids **[3, 132, 234]**.
- User Desktop index is **0-based** (unlike Create/Switch which use 1-based Desktop N).
- Production `CGSCopySpacesForWindows` mask is **7**.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

// canonicalType0Spaces is the locked experiment fixture: three user Desktops.
func canonicalType0Spaces() []SpaceInfoInput {
	return []SpaceInfoInput{
		{ID: 3, Type: 0},
		{ID: 132, Type: 0},
		{ID: 234, Type: 0},
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	// Shared defaults only; grouping/leaf Setups set Phase and fixtures.
	return nil
}
```
