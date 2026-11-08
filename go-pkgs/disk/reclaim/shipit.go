package reclaim

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// shipItEntry is one Library/Caches/*.ShipIt directory inventory row.
type shipItEntry struct {
	Name string
	Size int64
	Hint string
}

// shipItInventory is the dry-run report of ShipIt updater caches (never deleted).
type shipItInventory struct {
	Entries []shipItEntry
	Total   int64
}

// measureShipIt discovers directories under {Home}/Library/Caches whose names
// end with ".ShipIt". Files (even if named *.ShipIt) are ignored. Sorted by name.
func measureShipIt(home string) shipItInventory {
	caches := filepath.Join(home, "Library", "Caches")
	dirents, err := os.ReadDir(caches)
	if err != nil {
		return shipItInventory{}
	}

	var names []string
	sizes := make(map[string]int64)
	for _, e := range dirents {
		name := e.Name()
		if !strings.HasSuffix(name, ".ShipIt") {
			continue
		}
		// Only directories qualify (a file named *.ShipIt is noise).
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(caches, name)
		sz := pathSize(p)
		names = append(names, name)
		sizes[name] = sz
	}
	sort.Strings(names)

	var inv shipItInventory
	for _, name := range names {
		sz := sizes[name]
		inv.Entries = append(inv.Entries, shipItEntry{
			Name: name,
			Size: sz,
			Hint: shipItAppHint(name),
		})
		inv.Total += sz
	}
	return inv
}

// shipItAppHint returns a best-effort "update from inside" app label for basename.
func shipItAppHint(basename string) string {
	low := strings.ToLower(basename)
	switch {
	case strings.Contains(low, "vscode"):
		return "Visual Studio Code (Check for Updates)"
	case strings.Contains(low, "seatalk"):
		return "SeaTalk"
	case strings.Contains(low, "todesktop"):
		return "Cursor / ToDesktop app"
	case strings.Contains(low, "kiro"):
		return "Kiro"
	default:
		return "update the app that owns this bundle id"
	}
}

// printShipItSection prints the dry-run ShipIt inventory and fixed annotation.
// Sizes are never included in TotalFreed / default would-free by the caller.
func printShipItSection(w io.Writer, use bool, inv shipItInventory) {
	fmt.Fprintln(w, gray(use, "ShipIt updater caches (Library/Caches/*.ShipIt):"))
	if len(inv.Entries) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, e := range inv.Entries {
			fmt.Fprintf(w, "  %s    %s (%d B)\n", e.Name, humanSize(e.Size), e.Size)
			fmt.Fprintf(w, "    → Update from inside: %s\n", e.Hint)
		}
		fmt.Fprintf(w, "  total: %s (%d B)\n", humanSize(inv.Total), inv.Total)
	}
	fmt.Fprintln(w, "note: ShipIt caches are not auto-deleted by clear-tmp.")
	fmt.Fprintln(w, "      Free space by updating the corresponding app from its own UI;")
	fmt.Fprintln(w, "      leftover logs/state after update are usually tiny.")
	if inv.Total > 0 {
		fmt.Fprintf(w, "Potential free via app UI (ShipIt): %s (%d B)\n", humanSize(inv.Total), inv.Total)
	}
}
