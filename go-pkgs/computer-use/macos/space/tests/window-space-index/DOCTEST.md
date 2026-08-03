# macOS Space — window → 0-based Desktop index (SkyLight)

## Version
0.0.2

Classic TDD (P1) for a reusable API in
`github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space` that maps a
CGWindowNumber / iTerm window id to a **0-based** user Desktop index via
SkyLight private APIs. **First display only**, **type==0** spaces only.
No live WindowServer required: L2 leaves inject ManagedDisplaySpaces +
CopySpacesForWindows fixtures. No L3 e2e.

**Status:** RED until implementer lands the public API (build fail or assert fail).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies a CG/iTerm `windowID` and optional inject hooks.
- **SpaceIndexForWindow** — public resolver: window id → 0-based Desktop index.
- **ManagedDisplaySpaces** — SkyLight `CGSCopyManagedDisplaySpaces` (or inject):
  ordered monitors; each has a `Spaces` list with `id` + `type`.
- **CopySpacesForWindows** — SkyLight `CGSCopySpacesForWindows(cid, mask=7, [id])`
  (or inject): returns space id(s) for the window.
- **UserSpaceIndex** — pure map: type==0 spaces on the **first** display, list
  order → dense indices `0 … n-1`.
- **Platform gate** — non-darwin → existing `ErrUnsupportedPlatform`.

### Behaviors

- Build type-0 index from first display Spaces only; skip `type != 0`.
- Look up window’s space id(s) in that map → 0-based index.
- Unknown / non-type-0 / empty CopySpaces result → **detectable error**
  (`ErrNotUserSpace` or `ErrSpaceNotFound`), never silent `0`.
- Production path uses mask **7** for `CGSCopySpacesForWindows`.
- Injectable opts so CI never needs live WindowServer / Accessibility.

## Decision Tree

```
window-space-index/
├── pure-map/                         [Phase=pure-* ~L1]
│   ├── build-type0-dense/            type0 filter + dense 0..n-1 map
│   ├── resolve-known/                known space id → index
│   └── resolve-unknown/              unknown id → error
├── space-index/                      [Phase=space-index-for-window ~L2]
│   ├── found/
│   │   ├── first-space/              window → first type0 id → 0
│   │   ├── middle-space/             window → 132 in [3,132,234] → 1
│   │   ├── last-space/               window → last type0 → last index
│   │   ├── multi-display-first-only/ indices from first display only
│   │   └── dense-skip-non-type0/     type!=0 skipped; dense over type0
│   ├── error/
│   │   ├── non-user-space/           type4 / not in map → error
│   │   ├── empty-window-spaces/      CopySpaces empty → error
│   │   └── other-display-space/      space only on 2nd display → error
│   ├── contract/
│   │   └── uses-mask-7/              spy: mask arg == 7
│   └── platform/
│       └── non-darwin/               ErrUnsupportedPlatform
```

## Test Index

| Leaf | Phase | Description | Expect |
|------|-------|-------------|--------|
| `pure-map/build-type0-dense/` | pure-build-index | type0 → dense map; type!=0 omitted | RED |
| `pure-map/resolve-known/` | pure-resolve | resolve known id → index | RED |
| `pure-map/resolve-unknown/` | pure-resolve | unknown id → error | RED |
| `space-index/found/first-space/` | space-index-for-window | window on first user space → 0 | RED |
| `space-index/found/middle-space/` | space-index-for-window | window on 132 → 1 | RED |
| `space-index/found/last-space/` | space-index-for-window | window on last user space → last | RED |
| `space-index/found/multi-display-first-only/` | space-index-for-window | multi-display fixture; first display only | RED |
| `space-index/found/dense-skip-non-type0/` | space-index-for-window | mixed types; dense type0 indices | RED |
| `space-index/error/non-user-space/` | space-index-for-window | not in type0 map → error | RED |
| `space-index/error/empty-window-spaces/` | space-index-for-window | empty CopySpaces → error | RED |
| `space-index/error/other-display-space/` | space-index-for-window | id only on 2nd display → error | RED |
| `space-index/contract/uses-mask-7/` | space-index-for-window | injectable hook sees mask 7 | RED |
| `space-index/platform/non-darwin/` | space-index-for-window | non-darwin → ErrUnsupportedPlatform | RED |

## How to Run

```sh
# from go-pkgs module root (wrk path or main repo replace)
cd external/dot-pkgs-master-2026-07-31/go-pkgs   # or module root with this tree

doctest vet ./computer-use/macos/space/tests/window-space-index
doctest test ./computer-use/macos/space/tests/window-space-index
```

Expect **RED**: missing `SpaceIndexForWindow` / pure helpers / error sentinels
until implementer lands them (compile fail or failing asserts).

### Intended public API (implementer pins names)

```go
// SpaceInfo is one entry from managed display Spaces.
type SpaceInfo struct {
    ID   uint64
    Type int // 0 = user Desktop
}

// DisplaySpaces is one monitor's Spaces list (order preserved).
type DisplaySpaces struct {
    Spaces []SpaceInfo
}

// WindowSpaceOption configures SpaceIndexForWindow (injection for tests).
type WindowSpaceOption func(*windowSpaceConfig)

// WithManagedDisplays injects CGSCopyManagedDisplaySpaces result.
// First slice element is the only display used for indexing.
func WithManagedDisplays(displays []DisplaySpaces) WindowSpaceOption

// WithWindowSpaceIDs injects the space id list for the window
// (as if CGSCopySpacesForWindows returned them).
func WithWindowSpaceIDs(spaceIDs []uint64) WindowSpaceOption

// WithCopySpacesForWindows injects the CopySpaces hook (mask must be 7 in prod).
func WithCopySpacesForWindows(fn func(mask int, windowIDs []uint64) ([]uint64, error)) WindowSpaceOption

// WithPlatformGOOS injects platform check (parallel-safe; prefer over SetGOOSForTest).
func WithPlatformGOOS(goos string) WindowSpaceOption

// SpaceIndexForWindow returns 0-based user Desktop index for windowID.
// First display only. type!=0 / unknown / empty → detectable error.
func SpaceIndexForWindow(windowID uint64, opts ...WindowSpaceOption) (int, error)

// BuildUserSpaceIndex maps type==0 space IDs to dense 0-based indices (list order).
func BuildUserSpaceIndex(spaces []SpaceInfo) map[uint64]int

// ResolveWindowSpaceIndex looks up window space id(s) in a user-space index.
func ResolveWindowSpaceIndex(index map[uint64]int, windowSpaceIDs []uint64) (int, error)

var (
    // ErrNotUserSpace: window's space id is not a type-0 user Desktop on first display.
    ErrNotUserSpace error
    // ErrSpaceNotFound: no space ids for window (empty CopySpaces result).
    ErrSpaceNotFound error
)
// ErrUnsupportedPlatform already exists.
```

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

// SpaceInfoInput mirrors space.SpaceInfo for leaf fixtures.
type SpaceInfoInput struct {
	ID   uint64
	Type int
}

// DisplayInput is one monitor fixture (Spaces list).
type DisplayInput struct {
	Spaces []SpaceInfoInput
}

// Request selects Phase and fixtures for pure map or SpaceIndexForWindow.
type Request struct {
	// Phase:
	//   pure-build-index | pure-resolve | space-index-for-window
	Phase string

	// Spaces is a single-display Spaces list (used when Displays is empty).
	Spaces []SpaceInfoInput
	// Displays is multi-monitor ManagedDisplaySpaces (first = primary).
	Displays []DisplayInput

	// WindowID is the CG/iTerm window number under test.
	WindowID uint64
	// WindowSpaceIDs is the injected CopySpacesForWindows result.
	WindowSpaceIDs []uint64
	// EmptyWindowSpaces forces an empty CopySpaces result (even if WindowSpaceIDs nil).
	EmptyWindowSpaces bool

	// ResolveSpaceIDs is pure-resolve lookup input (defaults to WindowSpaceIDs).
	ResolveSpaceIDs []uint64

	// ForceGOOS injects platform via WithPlatformGOOS (e.g. "linux").
	ForceGOOS string
	// CaptureMask records mask passed to WithCopySpacesForWindows spy.
	CaptureMask bool
}

// Response holds pure-map and SpaceIndexForWindow results.
type Response struct {
	Index            int
	IndexMap         map[uint64]int
	CapturedMask     int
	CopySpacesCalled bool
}

// Run calls the intended package API. Classic TDD: RED until symbols exist.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = t
	_ = d
	resp := &Response{}

	switch req.Phase {
	case "pure-build-index":
		resp.IndexMap = space.BuildUserSpaceIndex(toSpaceInfos(req.Spaces))
		return resp, nil

	case "pure-resolve":
		idx := space.BuildUserSpaceIndex(toSpaceInfos(req.Spaces))
		ids := req.ResolveSpaceIDs
		if len(ids) == 0 {
			ids = req.WindowSpaceIDs
		}
		n, err := space.ResolveWindowSpaceIndex(idx, ids)
		resp.Index = n
		return resp, err

	case "space-index-for-window":
		opts := buildWindowSpaceOpts(req, resp)
		n, err := space.SpaceIndexForWindow(req.WindowID, opts...)
		resp.Index = n
		return resp, err

	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

func toSpaceInfos(in []SpaceInfoInput) []space.SpaceInfo {
	out := make([]space.SpaceInfo, len(in))
	for i, s := range in {
		out[i] = space.SpaceInfo{ID: s.ID, Type: s.Type}
	}
	return out
}

func toDisplaySpaces(in []DisplayInput) []space.DisplaySpaces {
	out := make([]space.DisplaySpaces, len(in))
	for i, d := range in {
		out[i] = space.DisplaySpaces{Spaces: toSpaceInfos(d.Spaces)}
	}
	return out
}

// buildWindowSpaceOpts assembles injectable WindowSpaceOption values.
// Pins WithManagedDisplays / WithWindowSpaceIDs / WithCopySpacesForWindows /
// WithPlatformGOOS as the test contract.
func buildWindowSpaceOpts(req *Request, resp *Response) []space.WindowSpaceOption {
	var opts []space.WindowSpaceOption

	if req.ForceGOOS != "" {
		opts = append(opts, space.WithPlatformGOOS(req.ForceGOOS))
	}

	displays := req.Displays
	if len(displays) == 0 && len(req.Spaces) > 0 {
		displays = []DisplayInput{{Spaces: req.Spaces}}
	}
	if len(displays) > 0 {
		opts = append(opts, space.WithManagedDisplays(toDisplaySpaces(displays)))
	}

	if req.CaptureMask {
		spaceIDs := append([]uint64(nil), req.WindowSpaceIDs...)
		opts = append(opts, space.WithCopySpacesForWindows(func(mask int, windowIDs []uint64) ([]uint64, error) {
			resp.CopySpacesCalled = true
			resp.CapturedMask = mask
			_ = windowIDs
			if req.EmptyWindowSpaces {
				return nil, nil
			}
			return spaceIDs, nil
		}))
	} else if req.EmptyWindowSpaces {
		opts = append(opts, space.WithWindowSpaceIDs(nil))
	} else if req.WindowSpaceIDs != nil {
		opts = append(opts, space.WithWindowSpaceIDs(append([]uint64(nil), req.WindowSpaceIDs...)))
	}

	return opts
}
```
