//go:build !darwin

package space

import "fmt"

func liveManagedDisplaySpacesImpl() ([]DisplaySpaces, error) {
	return nil, fmt.Errorf("space: managed display spaces require macOS (use WithManagedDisplays for tests)")
}

func liveCopySpacesForWindowsImpl(mask int, windowIDs []uint64) ([]uint64, error) {
	_ = mask
	_ = windowIDs
	return nil, fmt.Errorf("space: CopySpacesForWindows requires macOS (use WithWindowSpaceIDs / WithCopySpacesForWindows for tests)")
}
