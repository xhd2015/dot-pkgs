// Package measure reports byte sizes of files, directory trees and glob patterns.
package measure

import (
	"os"
	"path/filepath"
)

// pathSize returns the total size of a file or directory tree.
// Missing paths return 0. Directories include all nested file sizes.
func PathSize(path string) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}
	var total int64
	_ = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi == nil || fi.IsDir() {
			return nil
		}
		total += fi.Size()
		return nil
	})
	return total
}

// globSize sums sizes of all paths matching pattern (filepath.Glob).
// Missing matches → 0.
func GlobSize(pattern string) int64 {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return 0
	}
	var total int64
	for _, m := range matches {
		total += PathSize(m)
	}
	return total
}

// filesSize sums sizes of the given paths; missing files contribute 0.
func FilesSize(paths ...string) int64 {
	var total int64
	for _, p := range paths {
		total += PathSize(p)
	}
	return total
}
