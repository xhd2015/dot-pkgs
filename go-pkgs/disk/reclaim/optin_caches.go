package reclaim

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	stepMSPlaywright     = "ms-playwright"
	stepOpenCodeSnapshot = "opencode-snapshot"
)

func msPlaywrightPath(home string) string {
	return filepath.Join(home, "Library", "Caches", "ms-playwright")
}

func openCodeSnapshotPath(home string) string {
	return filepath.Join(home, ".local", "share", "opencode", "snapshot")
}

func measureMSPlaywright(opts Options) int64 {
	return pathSize(msPlaywrightPath(opts.Home))
}

func measureOpenCodeSnapshot(opts Options) int64 {
	return pathSize(openCodeSnapshotPath(opts.Home))
}

func applyMSPlaywright(opts Options) error {
	p := msPlaywrightPath(opts.Home)
	if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func applyOpenCodeSnapshot(opts Options) error {
	p := openCodeSnapshotPath(opts.Home)
	if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// printOptInCacheSection prints a dry-run inventory row for an opt-in cache path.
func printOptInCacheSection(w io.Writer, use bool, title, path string, size int64, included bool, flag string) {
	fmt.Fprintln(w, gray(use, title+":"))
	if size <= 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		status := "not deleted by default"
		if included {
			status = "would delete / included"
		}
		fmt.Fprintf(w, "  %s  %s (%d B)  [%s]\n", path, humanSize(size), size, status)
	}
	fmt.Fprintf(w, "note: not deleted by default; pass %s to include.\n", flag)
}
