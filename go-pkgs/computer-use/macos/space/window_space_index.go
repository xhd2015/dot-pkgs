package space

import (
	"errors"
	"fmt"
)

// SpaceInfo is one entry from managed display Spaces (CGS / inject).
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

// copySpacesMask is the production mask for CGSCopySpacesForWindows
// (locked experiment / contract: mask 7).
const copySpacesMask = 7

var (
	// ErrNotUserSpace: window's space id is not a type-0 user Desktop on the
	// first display (or is type!=0 / only present on another display).
	ErrNotUserSpace = errors.New("space: window space is not a user Desktop on first display")
	// ErrSpaceNotFound: no space ids for window (empty CopySpaces result).
	ErrSpaceNotFound = errors.New("space: no space ids for window")
)

type windowSpaceConfig struct {
	// platformGOOS, when set, overrides effectiveGOOS for this call (parallel-safe).
	platformGOOS    string
	hasPlatformGOOS bool

	// managedDisplays injects CGSCopyManagedDisplaySpaces result.
	managedDisplays    []DisplaySpaces
	hasManagedDisplays bool

	// windowSpaceIDs injects the CopySpacesForWindows result for the window.
	windowSpaceIDs    []uint64
	hasWindowSpaceIDs bool

	// copySpaces injects the CopySpacesForWindows hook (mask must be 7 in prod).
	copySpaces func(mask int, windowIDs []uint64) ([]uint64, error)
}

// WithManagedDisplays injects CGSCopyManagedDisplaySpaces result.
// First slice element is the only display used for indexing.
func WithManagedDisplays(displays []DisplaySpaces) WindowSpaceOption {
	return func(c *windowSpaceConfig) {
		c.managedDisplays = displays
		c.hasManagedDisplays = true
	}
}

// WithWindowSpaceIDs injects the space id list for the window
// (as if CGSCopySpacesForWindows returned them).
// A nil or empty list is treated as an empty CopySpaces result.
func WithWindowSpaceIDs(spaceIDs []uint64) WindowSpaceOption {
	return func(c *windowSpaceConfig) {
		c.windowSpaceIDs = spaceIDs
		c.hasWindowSpaceIDs = true
	}
}

// WithCopySpacesForWindows injects the CopySpaces hook.
// Production always passes mask 7; injects receive that mask for contract tests.
func WithCopySpacesForWindows(fn func(mask int, windowIDs []uint64) ([]uint64, error)) WindowSpaceOption {
	return func(c *windowSpaceConfig) {
		c.copySpaces = fn
	}
}

// WithPlatformGOOS injects platform check (parallel-safe; prefer over SetGOOSForTest).
func WithPlatformGOOS(goos string) WindowSpaceOption {
	return func(c *windowSpaceConfig) {
		c.platformGOOS = goos
		c.hasPlatformGOOS = true
	}
}

// BuildUserSpaceIndex maps type==0 space IDs to dense 0-based indices in list order.
// Entries with type != 0 are omitted; indices are dense over type-0 only.
func BuildUserSpaceIndex(spaces []SpaceInfo) map[uint64]int {
	out := make(map[uint64]int)
	n := 0
	for _, s := range spaces {
		if s.Type != 0 {
			continue
		}
		out[s.ID] = n
		n++
	}
	return out
}

// ResolveWindowSpaceIndex looks up window space id(s) in a user-space index.
// Empty windowSpaceIDs → ErrSpaceNotFound.
// Non-empty ids with no match in index → ErrNotUserSpace.
// When multiple ids are present, the first id found in the index wins.
func ResolveWindowSpaceIndex(index map[uint64]int, windowSpaceIDs []uint64) (int, error) {
	if len(windowSpaceIDs) == 0 {
		return 0, ErrSpaceNotFound
	}
	for _, id := range windowSpaceIDs {
		if n, ok := index[id]; ok {
			return n, nil
		}
	}
	return 0, ErrNotUserSpace
}

// SpaceIndexForWindow returns the 0-based user Desktop index for windowID on
// the first managed display (type==0 Spaces only).
//
// Injectable opts (WithManagedDisplays, WithWindowSpaceIDs,
// WithCopySpacesForWindows, WithPlatformGOOS) allow L2 tests without live
// WindowServer. Production path uses mask 7 for CGSCopySpacesForWindows.
//
// Errors:
//   - ErrUnsupportedPlatform when GOOS is not darwin
//   - ErrSpaceNotFound when CopySpaces returns no space ids
//   - ErrNotUserSpace when the window's space is not a first-display type-0 Desktop
func SpaceIndexForWindow(windowID uint64, opts ...WindowSpaceOption) (int, error) {
	cfg := &windowSpaceConfig{}
	for _, o := range opts {
		if o != nil {
			o(cfg)
		}
	}

	goos := effectiveGOOS()
	if cfg.hasPlatformGOOS {
		goos = cfg.platformGOOS
	}
	if goos != "darwin" {
		return 0, ErrUnsupportedPlatform
	}

	displays, err := resolveManagedDisplays(cfg)
	if err != nil {
		return 0, err
	}
	if len(displays) == 0 {
		return 0, fmt.Errorf("space: no managed displays: %w", ErrSpaceNotFound)
	}

	// First display only.
	index := BuildUserSpaceIndex(displays[0].Spaces)

	spaceIDs, err := resolveWindowSpaceIDs(cfg, windowID)
	if err != nil {
		return 0, err
	}
	return ResolveWindowSpaceIndex(index, spaceIDs)
}

func resolveManagedDisplays(cfg *windowSpaceConfig) ([]DisplaySpaces, error) {
	if cfg.hasManagedDisplays {
		return cfg.managedDisplays, nil
	}
	return liveManagedDisplaySpaces()
}

func resolveWindowSpaceIDs(cfg *windowSpaceConfig, windowID uint64) ([]uint64, error) {
	// Prefer explicit CopySpaces hook (contract / spy path) over static IDs.
	if cfg.copySpaces != nil {
		return cfg.copySpaces(copySpacesMask, []uint64{windowID})
	}
	if cfg.hasWindowSpaceIDs {
		return cfg.windowSpaceIDs, nil
	}
	return liveCopySpacesForWindows(copySpacesMask, []uint64{windowID})
}

// liveManagedDisplaySpaces is the production ManagedDisplaySpaces path.
// Overridden on darwin with SkyLight when available; default reports unsupported
// so injectable tests never require WindowServer.
func liveManagedDisplaySpaces() ([]DisplaySpaces, error) {
	return liveManagedDisplaySpacesImpl()
}

// liveCopySpacesForWindows is the production CGSCopySpacesForWindows path.
func liveCopySpacesForWindows(mask int, windowIDs []uint64) ([]uint64, error) {
	return liveCopySpacesForWindowsImpl(mask, windowIDs)
}
