//go:build darwin && !cgo

package space

import "fmt"

// countDesktopsPrivateAPI is a pure-Go stub for darwin builds without cgo
// (e.g. GOOS=darwin GOARCH=amd64 cross-compile with CGO_ENABLED=0).
// The real SkyLight implementation is in count_private_darwin.go (darwin && cgo).
func countDesktopsPrivateAPI() (int, error) {
	return 0, fmt.Errorf("space: private count API requires cgo")
}
