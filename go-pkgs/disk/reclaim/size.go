package reclaim

import "github.com/xhd2015/dot-pkgs/go-pkgs/disk/measure"

// pathSize returns the total size of a file or directory tree.
func pathSize(path string) int64 { return measure.PathSize(path) }

// globSize sums sizes of all paths matching pattern.
func globSize(pattern string) int64 { return measure.GlobSize(pattern) }

// filesSize sums sizes of the given paths; missing files contribute 0.
func filesSize(paths ...string) int64 { return measure.FilesSize(paths...) }
