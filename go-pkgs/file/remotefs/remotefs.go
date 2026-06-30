package remotefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsRemoteBackedPath reports whether path lives under a macOS cloud-sync root
// (for example ~/Library/CloudStorage/<provider>/). Detection uses a path
// heuristic after normalization; stat errors are returned to the caller.
func IsRemoteBackedPath(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	abs = filepath.Clean(abs)
	if _, err := os.Stat(abs); err != nil {
		return false, err
	}
	return strings.Contains(abs, cloudStorageMarker()), nil
}

func cloudStorageMarker() string {
	return string(filepath.Separator) + "Library" + string(filepath.Separator) + "CloudStorage" + string(filepath.Separator)
}