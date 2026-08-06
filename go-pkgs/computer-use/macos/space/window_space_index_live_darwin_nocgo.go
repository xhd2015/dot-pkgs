//go:build darwin && !cgo

package space

import "fmt"

// Stubs for darwin builds without cgo so the package still links when
// cross-compiling (CGO_ENABLED=0). Live SkyLight implementations live in
// window_space_index_live_darwin.go (darwin && cgo).

func liveManagedDisplaySpacesImpl() ([]DisplaySpaces, error) {
	return nil, fmt.Errorf("space: managed display spaces require cgo")
}

func liveCopySpacesForWindowsImpl(mask int, windowIDs []uint64) ([]uint64, error) {
	_ = mask
	_ = windowIDs
	return nil, fmt.Errorf("space: CopySpacesForWindows requires cgo")
}
