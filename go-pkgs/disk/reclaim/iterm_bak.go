package reclaim

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const stepITermBakApps = "iterm-bak-apps"

// itermBakEntry is one backup .app bundle under Applications.
type itermBakEntry struct {
	Path    string
	Name    string
	Size    int64
	Running bool
}

// itermBakInventory is the discovered set of iTerm/iTerm2 bak apps.
type itermBakInventory struct {
	Entries   []itermBakEntry
	Deletable int64 // sum of sizes where !Running
	Total     int64 // all matching bak apps
	Skipped   int   // running count
}

// defaultApplicationsDirs returns {Home}/Applications and /Applications.
func defaultApplicationsDirs(home string) []string {
	return []string{
		filepath.Join(home, "Applications"),
		"/Applications",
	}
}

// applicationsDirs resolves Options.ApplicationsDirs or the default pair.
func applicationsDirs(opts Options) []string {
	if len(opts.ApplicationsDirs) > 0 {
		return append([]string(nil), opts.ApplicationsDirs...)
	}
	return defaultApplicationsDirs(opts.Home)
}

// isITermBakName reports whether basename is an iTerm/iTerm2 backup app name.
// Matches (case-insensitive): iTerm.app.bak-*, iTerm.app.back-*,
// iTerm2.app.bak-*, iTerm2.app.back-*.
func isITermBakName(name string) bool {
	lower := strings.ToLower(name)
	for _, p := range []string{
		"iterm.app.bak-",
		"iterm.app.back-",
		"iterm2.app.bak-",
		"iterm2.app.back-",
	} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// isAppRunning reports whether any process is executing from under appPath.
// opts.IsAppRunning overrides; nil uses ps(1). On ps failure, returns true
// (fail closed: skip delete when we cannot verify).
func isAppRunning(opts Options, appPath string) bool {
	if opts.IsAppRunning != nil {
		return opts.IsAppRunning(appPath)
	}
	return defaultIsAppRunning(appPath)
}

func defaultIsAppRunning(appPath string) bool {
	appPath = filepath.Clean(appPath)
	out, err := exec.Command("ps", "-ax", "-o", "command=").Output()
	if err != nil {
		return true
	}
	prefix := appPath + string(os.PathSeparator)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip leading argv noise: command column is usually absolute path + args.
		cmdPath := line
		if i := strings.IndexByte(line, ' '); i >= 0 {
			cmdPath = line[:i]
		}
		if cmdPath == appPath || strings.HasPrefix(cmdPath, prefix) {
			return true
		}
		// Also match if the full command line embeds the bundle path (helpers).
		if strings.Contains(line, prefix) || strings.Contains(line, appPath+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// listITermBakApps walks Applications dirs for matching bak bundles.
func listITermBakApps(opts Options) itermBakInventory {
	var inv itermBakInventory
	seen := make(map[string]struct{})
	for _, dir := range applicationsDirs(opts) {
		ents, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range ents {
			name := e.Name()
			if !isITermBakName(name) {
				continue
			}
			// Prefer directories (real .app bundles); ignore plain files.
			if !e.IsDir() {
				// ReadDir follows symlinks as the target type on some systems;
				// also accept symlinks to dirs via Stat.
				info, err := e.Info()
				if err != nil {
					continue
				}
				if !info.IsDir() {
					// Check if symlink to directory.
					p := filepath.Join(dir, name)
					st, err := os.Stat(p)
					if err != nil || !st.IsDir() {
						continue
					}
				}
			}
			p := filepath.Join(dir, name)
			abs, err := filepath.Abs(p)
			if err != nil {
				abs = p
			}
			abs = filepath.Clean(abs)
			if _, ok := seen[abs]; ok {
				continue
			}
			seen[abs] = struct{}{}
			sz := pathSize(abs)
			running := isAppRunning(opts, abs)
			inv.Entries = append(inv.Entries, itermBakEntry{
				Path:    abs,
				Name:    name,
				Size:    sz,
				Running: running,
			})
			inv.Total += sz
			if running {
				inv.Skipped++
			} else {
				inv.Deletable += sz
			}
		}
	}
	sort.Slice(inv.Entries, func(i, j int) bool {
		return inv.Entries[i].Path < inv.Entries[j].Path
	})
	return inv
}

// measureITermBakApps returns the would-free / before size (non-running only).
func measureITermBakApps(opts Options) int64 {
	return listITermBakApps(opts).Deletable
}

// applyITermBakApps removes non-running bak apps; warns about skipped running ones.
func applyITermBakApps(opts Options) error {
	inv := listITermBakApps(opts)
	for _, e := range inv.Entries {
		if e.Running {
			warnf(opts.Stderr, "skipping running iTerm bak app: %s", e.Path)
			continue
		}
		if err := os.RemoveAll(e.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", e.Path, err)
		}
	}
	return nil
}

// printITermBakSection prints dry-run inventory of bak apps with delete/skip status.
func printITermBakSection(w io.Writer, use bool, inv itermBakInventory, dryRun bool) {
	fmt.Fprintln(w, gray(use, "iTerm bak apps ({Home}/Applications + /Applications):"))
	if len(inv.Entries) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, e := range inv.Entries {
			status := "would delete"
			if !dryRun {
				status = "deleted"
			}
			if e.Running {
				status = "skipped: running"
			}
			fmt.Fprintf(w, "  %s  %s (%d B)  [%s]\n", e.Name, humanSize(e.Size), e.Size, status)
			fmt.Fprintf(w, "    %s\n", e.Path)
		}
		fmt.Fprintf(w, "  deletable: %s (%d B); skipped running: %d\n",
			humanSize(inv.Deletable), inv.Deletable, inv.Skipped)
	}
	fmt.Fprintln(w, "note: live iTerm.app / iTerm2.app are never deleted; only *.bak-* / *.back-* backups.")
}
