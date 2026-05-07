package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/xhd2015/less-gen/flags"
	llsconfig "github.com/xhd2015/lls/config"
)

// History tracks movement history for each file.
// Key: original absolute path, Value: list of locations (oldest first).
type History map[string][]string

const help = `
Usage: mvd [OPTIONS] SRC [DST]
       mvd rm [-f] DIR
       mvd rebase DIR NEW-DIR

Move a file/directory and track its location history.

Commands:
  mvd SRC DST          Move SRC to DST (tracked)
  mvd --add DIR        Add DIR to the record file without moving it
  mvd rm [-f] DIR      Remove the exact recorded entry for DIR
  mvd rebase DIR NEW-DIR
                       Change the entry base to NEW-DIR
  mvd --list [SRC]     Show location history (all if SRC omitted)
  mvd --back SRC       Move back by current path, root path, or unique root basename
  mvd --clear SRC      Clear movement history for SRC

Options:
  --add                Add a DIR to history without moving it
  --list               Show location history
  --back               Move back to previous location
  --clear              Clear movement history for a specific file
  -f, --force          Force removal for mvd rm and clear its histories
  -h, --help           Show this help message
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "mvd: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "rm":
			return runRemove(args[1:])
		case "rebase":
			if len(args) != 3 {
				return fmt.Errorf("usage: mvd rebase DIR NEW-DIR")
			}
			return cmdRebase(args[1], args[2])
		}
	}

	var add, list, back, clear bool
	args, err := flags.Bool("--add", &add).
		Bool("--list", &list).
		Bool("--back", &back).
		Bool("--clear", &clear).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if add {
		if len(args) != 1 {
			return fmt.Errorf("usage: mvd --add DIR")
		}
		return cmdAdd(args[0])
	}
	if clear {
		if len(args) < 1 {
			return fmt.Errorf("usage: mvd --clear SRC")
		}
		return cmdClear(args[0])
	}
	if list {
		if len(args) > 0 {
			return cmdList(args[0])
		}
		return cmdListAll()
	}
	if back {
		if len(args) < 1 {
			return fmt.Errorf("usage: mvd --back SRC")
		}
		return cmdBack(args[0])
	}

	if len(args) < 2 {
		fmt.Print(help)
		return nil
	}
	return cmdMove(args[0], args[1])
}

func runRemove(args []string) error {
	var force bool
	var dir string

	for _, arg := range args {
		switch arg {
		case "-f", "--force":
			force = true
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return fmt.Errorf("usage: mvd rm [-f] DIR")
			}
			if dir != "" {
				return fmt.Errorf("usage: mvd rm [-f] DIR")
			}
			dir = arg
		}
	}
	if dir == "" {
		return fmt.Errorf("usage: mvd rm [-f] DIR")
	}
	return cmdRemove(dir, force)
}

func cmdAdd(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	if _, locations := findEntry(hist, absDir); locations != nil {
		fmt.Printf("hint: %s is already recorded, nothing added\n", displayPath(absDir))
		return nil
	}

	hist[absDir] = []string{absDir}
	fmt.Printf("added: %s\n", displayPath(absDir))
	return saveHistory(hist)
}

func cmdRemove(dir string, force bool) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	locations, ok := hist[absDir]
	if !ok {
		fmt.Printf("hint: no recorded entry for %s\n", displayPath(absDir))
		return nil
	}

	if len(locations) > 1 {
		if !force {
			return fmt.Errorf("%s has movement history:\n  use `mvd rm -f %s`\n  to clear it", displayPath(absDir), absDir)
		}
		fmt.Printf("hint: removing %s will clear %d history entries\n", displayPath(absDir), len(locations)-1)
	}

	delete(hist, absDir)
	fmt.Printf("removed: %s\n", displayPath(absDir))
	return saveHistory(hist)
}

func cmdRebase(dir, newDir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	absNewDir, err := filepath.Abs(newDir)
	if err != nil {
		return fmt.Errorf("resolve new-dir: %w", err)
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations := findEntry(hist, absDir)
	if locations == nil {
		return fmt.Errorf("no recorded entry for %s", absDir)
	}

	if otherKey, otherLocations := findEntry(hist, absNewDir); otherLocations != nil && otherKey != origKey {
		return fmt.Errorf("%s is already recorded in another entry", absNewDir)
	}

	if locations[0] == absNewDir {
		fmt.Printf("base unchanged: %s\n", displayPath(absNewDir))
		return nil
	}

	oldBase := locations[0]
	updated := []string{absNewDir}
	if oldBase != absNewDir && !containsPath(locations[1:], oldBase) {
		updated = append(updated, oldBase)
	}
	updated = append(updated, locations[1:]...)

	delete(hist, origKey)
	hist[absNewDir] = updated
	fmt.Printf("rebased: %s → %s\n", displayPath(origKey), displayPath(absNewDir))
	return saveHistory(hist)
}

func cmdClear(src string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations := findEntry(hist, absSrc)
	if locations == nil {
		fmt.Printf("no movement history for %s\n", displayPath(absSrc))
		return nil
	}

	delete(hist, origKey)
	fmt.Printf("cleared history for %s (%d entries)\n", displayPath(absSrc), len(locations))
	return saveHistory(hist)
}

func cmdMove(src, dst string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations, absSrc, err := resolveMoveSource(hist, src)
	if err != nil {
		return err
	}

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve dst: %w", err)
	}

	// If dst is a directory, move src into it
	info, err := os.Stat(absDst)
	if err == nil && info.IsDir() {
		absDst = filepath.Join(absDst, filepath.Base(absSrc))
	}

	if err := os.Rename(absSrc, absDst); err != nil {
		return fmt.Errorf("move: %w", err)
	}

	if locations == nil {
		origKey = absSrc
		locations = []string{absSrc, absDst}
	} else {
		locations = append(locations, absDst)
	}

	delete(hist, origKey)
	hist[locations[0]] = locations

	return saveHistory(hist)
}

func resolveMoveSource(hist History, src string) (string, []string, string, error) {
	if useRootBaseNameShortcut(src) {
		origKey, locations, err := findEntryByRootBaseName(hist, src)
		if err != nil {
			return "", nil, "", err
		}
		if locations != nil {
			if len(locations) == 0 {
				return "", nil, "", fmt.Errorf("empty mv history for %s", src)
			}
			return origKey, locations, locations[len(locations)-1], nil
		}
	}

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return "", nil, "", fmt.Errorf("resolve src: %w", err)
	}

	origKey, locations := findEntry(hist, absSrc)
	if locations == nil {
		return "", nil, absSrc, nil
	}
	if len(locations) == 0 {
		return "", nil, "", fmt.Errorf("empty mv history for %s", absSrc)
	}

	last := locations[len(locations)-1]
	root := locations[0]
	if absSrc != root && absSrc != last {
		return "", nil, "", fmt.Errorf("current position mismatch: expected %s at end of history, got %s", absSrc, last)
	}

	return origKey, locations, last, nil
}

func cmdListAll() error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		fmt.Println("(no mv history)")
		return nil
	}

	// Sort keys for stable output
	keys := make([]string, 0, len(hist))
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i > 0 {
			fmt.Println()
		}
		locations := hist[key]
		current := locations[len(locations)-1]
		fmt.Printf("%s\n", displayPath(key))
		// Skip locations[0] because it's identical to the header
		// (key). Printing it again just duplicates information.
		for _, loc := range locations[1:] {
			marker := "  "
			if loc == current {
				marker = "* "
			}
			fmt.Printf("  %s%s\n", marker, displayPath(loc))
		}
	}
	return nil
}

func cmdList(src string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	locations := findLocations(hist, absSrc)
	if locations == nil {
		fmt.Println("(no mv history)")
		return nil
	}

	for i, loc := range locations {
		marker := "  "
		if loc == absSrc {
			marker = "* "
		}
		if i == 0 {
			fmt.Printf("%s%s  (original)\n", marker, displayPath(loc))
		} else {
			fmt.Printf("%s%s\n", marker, displayPath(loc))
		}
	}
	return nil
}

func cmdBack(src string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations, err := resolveBackEntry(hist, src)
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		return fmt.Errorf("empty mv history for %s", src)
	}
	last := locations[len(locations)-1]
	if len(locations) <= 1 {
		fmt.Printf("nothing to move back for %s\n", displayPath(last))
		return nil
	}

	prev := locations[len(locations)-2]
	if err := os.Rename(last, prev); err != nil {
		return fmt.Errorf("move back: %w", err)
	}
	fmt.Printf("moved back: %s → %s\n", displayPath(last), displayPath(prev))

	locations = locations[:len(locations)-1]
	hist[origKey] = locations

	return saveHistory(hist)
}

func resolveBackEntry(hist History, src string) (string, []string, error) {
	if useRootBaseNameShortcut(src) {
		origKey, locations, err := findEntryByRootBaseName(hist, src)
		if err != nil {
			return "", nil, err
		}
		if locations != nil {
			return origKey, locations, nil
		}
	}

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return "", nil, fmt.Errorf("resolve: %w", err)
	}

	origKey, locations := findEntry(hist, absSrc)
	if locations == nil {
		return "", nil, fmt.Errorf("no mv history for %s", absSrc)
	}
	if len(locations) == 0 {
		return "", nil, fmt.Errorf("empty mv history for %s", absSrc)
	}

	last := locations[len(locations)-1]
	root := locations[0]
	if absSrc != root && absSrc != last {
		return "", nil, fmt.Errorf("current position mismatch: expected %s at end of history, got %s", absSrc, last)
	}

	return origKey, locations, nil
}

func useRootBaseNameShortcut(path string) bool {
	if !isBareBaseName(path) {
		return false
	}
	_, err := os.Lstat(path)
	return os.IsNotExist(err)
}

func isBareBaseName(path string) bool {
	return path != "." && path != ".." && filepath.Base(path) == path
}

func findEntryByRootBaseName(hist History, baseName string) (string, []string, error) {
	var matchedKey string
	var matchedLocations []string
	var matchedRoots []string

	for key, locations := range hist {
		if len(locations) == 0 {
			continue
		}
		root := locations[0]
		if filepath.Base(root) != baseName {
			continue
		}
		matchedRoots = append(matchedRoots, root)
		matchedKey = key
		matchedLocations = locations
	}

	if len(matchedRoots) > 1 {
		sort.Strings(matchedRoots)
		return "", nil, fmt.Errorf("ambiguous root basename %s matches multiple roots: %v", baseName, displayPaths(matchedRoots))
	}
	return matchedKey, matchedLocations, nil
}

var (
	displayConfigOnce sync.Once
	displayConfig     *llsconfig.Config
)

func displayPath(path string) string {
	cfg := loadDisplayConfig()
	if cfg == nil {
		return path
	}
	return llsconfig.CollapsePath(path, cfg.Envs)
}

func displayPaths(paths []string) []string {
	displayed := make([]string, len(paths))
	for i, path := range paths {
		displayed[i] = displayPath(path)
	}
	return displayed
}

func loadDisplayConfig() *llsconfig.Config {
	displayConfigOnce.Do(func() {
		file, err := llsconfig.DefaultFile(false)
		if err != nil {
			return
		}
		if _, err := os.Stat(file); err != nil {
			return
		}
		cfg, err := llsconfig.Load(file)
		if err != nil || len(cfg.Envs) == 0 {
			return
		}
		displayConfig = &cfg
	})
	return displayConfig
}

func findLocations(hist History, path string) []string {
	_, locs := findEntry(hist, path)
	return locs
}

func findEntry(hist History, path string) (string, []string) {
	if locs, ok := hist[path]; ok {
		return path, locs
	}
	for key, locs := range hist {
		for _, loc := range locs {
			if loc == path {
				return key, locs
			}
		}
		_ = key
	}
	return "", nil
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}

func historyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mvd", "history.json")
}

func loadHistory() (History, error) {
	p := historyPath()
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return make(History), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}
	var hist History
	if err := json.Unmarshal(data, &hist); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return mergeChains(hist), nil
}

// mergeChains stitches together entries whose tail matches another
// entry's head. Older versions of mvd stored A→B and B→C as two
// separate entries whenever the user ran `mvd B C`; this produces the
// unified chain A→B→C that the user actually moved around.
//
// The merge is repeated until no more chains can be joined, which
// handles arbitrary-length splits (A→B, B→C, C→D, ...). Duplicate
// adjacent locations introduced by the overlap at the seam are also
// collapsed.
func mergeChains(hist History) History {
	for {
		merged := false
		for key, locs := range hist {
			if len(locs) == 0 {
				continue
			}
			tail := locs[len(locs)-1]
			if tail == key {
				continue
			}
			next, ok := hist[tail]
			if !ok || len(next) == 0 {
				continue
			}
			combined := append([]string{}, locs...)
			// Skip next[0] because it duplicates tail.
			combined = append(combined, next[1:]...)
			hist[key] = combined
			delete(hist, tail)
			merged = true
			break
		}
		if !merged {
			return hist
		}
	}
}

func saveHistory(hist History) error {
	p := historyPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	data, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}
