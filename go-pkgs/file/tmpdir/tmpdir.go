// Package tmpdir selects temporary-directory parents suitable for shared local files.
package tmpdir

import (
	"os"
	"runtime"
)

const unixTmpDir = "/tmp"

// GetCommonTmpDir returns /tmp on non-Windows systems when it is a directory.
// It otherwise returns the operating system's default temporary directory.
func GetCommonTmpDir() string {
	return getCommonTmpDir(runtime.GOOS, os.Stat, os.TempDir)
}

func getCommonTmpDir(goos string, stat func(string) (os.FileInfo, error), tempDir func() string) string {
	if goos != "windows" {
		if info, err := stat(unixTmpDir); err == nil && info.IsDir() {
			return unixTmpDir
		}
	}
	return tempDir()
}
