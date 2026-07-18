//go:build !darwin

package space

import "fmt"

func countDesktopsPrivateAPI() (int, error) {
	return 0, fmt.Errorf("space: private count API is only available on macOS")
}
