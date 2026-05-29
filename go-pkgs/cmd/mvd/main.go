package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/xhd2015/less-gen/flags"
	llsconfig "github.com/xhd2015/lls/config"
)

// History tracks movement history for each file.
// Key: original absolute path, Value: list of locations (oldest first).
type History map[string][]string

const help = `
Usage: mvd [OPTIONS] SRC [DST]
       mvd --rm|--remove [-f] DIR
       mvd --add-alias ALIAS PROJECT
       mvd --rebase DIR NEW-DIR
       mvd --which SRC
       mvd --print SRC

Move a file/directory and track its location history.

Commands:
  mvd SRC DST          Move SRC to DST (tracked)
  mvd --add DIR        Add DIR to the record file without moving it
  mvd --add-alias ALIAS PROJECT
                       Set ALIAS to a tracked project
  mvd --rm|--remove [-f] DIR
                       Remove the exact recorded entry for DIR
  mvd --rebase DIR NEW-DIR
                       Change the entry base to NEW-DIR
  mvd --list [SRC]     Show location history (all if SRC omitted)
  mvd --which SRC      Show all move-source matches in resolution order
  mvd --back SRC       Move back by current path, root path, or unique root basename
  mvd --clear SRC      Clear movement history for SRC
  mvd --print SRC      Print shortened and full paths for SRC

Options:
  --add                Add a DIR to history without moving it
  --add-alias          Set an alias for a tracked project
  --rm, --remove       Remove the exact recorded entry for DIR
  --rebase             Change the entry base to NEW-DIR
  --list               Show location history
  --which              Show all move-source matches
  --back               Move back to previous location
  --clear              Clear movement history for a specific file
  -p, --print          Print shortened and full paths
  -f, --force          Force removal for mvd --rm and clear its histories
  -h, --help           Show this help message
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "mvd: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var add, remove, rebase, list, which, back, clear, print, force bool
	var addAlias string
	args, err := flags.Bool("--add", &add).
		String("--add-alias", &addAlias).
		Bool("--rm,--remove", &remove).
		Bool("--rebase", &rebase).
		Bool("--list", &list).
		Bool("--which", &which).
		Bool("--back", &back).
		Bool("--clear", &clear).
		Bool("-p,--print", &print).
		Bool("-f,--force", &force).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	modeCount := 0
	for _, enabled := range []bool{remove, add, addAlias != "", rebase, list, which, back, clear, print} {
		if enabled {
			modeCount++
		}
	}
	if modeCount > 1 {
		return fmt.Errorf("at most one of --rm, --add, --add-alias, --rebase, --list, --which, --back, --clear, --print can be specified")
	}

	if force && !remove {
		return fmt.Errorf("-f, --force requires --rm")
	}

	if add {
		if len(args) != 1 {
			return fmt.Errorf("usage: mvd --add DIR")
		}
		return cmdAdd(args[0])
	}
	if addAlias != "" {
		if len(args) != 1 {
			return fmt.Errorf("usage: mvd --add-alias ALIAS PROJECT")
		}
		return cmdAddAlias(addAlias, args[0])
	}
	if remove {
		if len(args) != 1 {
			return fmt.Errorf("usage: mvd --rm [-f] DIR")
		}
		return cmdRemove(args[0], force)
	}
	if rebase {
		if len(args) != 2 {
			return fmt.Errorf("usage: mvd --rebase DIR NEW-DIR")
		}
		return cmdRebase(args[0], args[1])
	}
	if clear {
		if len(args) < 1 {
			return fmt.Errorf("usage: mvd --clear SRC")
		}
		return cmdClear(args[0])
	}
	if print {
		if len(args) != 1 {
			return fmt.Errorf("usage: mvd --print SRC")
		}
		return cmdPrint(args[0])
	}
	if list {
		if len(args) > 0 {
			return cmdList(args[0])
		}
		return cmdListAll()
	}
	if which {
		if len(args) != 1 {
			return fmt.Errorf("usage: mvd --which SRC")
		}
		return cmdWhich(args[0])
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

func cmdAddAlias(alias, project string) error {
	if err := validateAliasName(alias); err != nil {
		return err
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	root, err := resolveAliasProject(hist, project)
	if err != nil {
		return err
	}

	aliases, err := loadAliases()
	if err != nil {
		return err
	}
	aliases[alias] = root
	fmt.Printf("alias: %s -> %s\n", alias, displayPath(root))
	return saveAliases(aliases)
}

func cmdAdd(dir string) error {
	absDir, err := resolveAddPath(dir)
	if err != nil {
		return err
	}

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	if isRecordedPath(hist, absDir) {
		fmt.Printf("hint: %s is already recorded, nothing added\n", displayPath(absDir))
		return nil
	}

	hist[absDir] = []string{absDir}
	fmt.Printf("added: %s\n", displayPath(absDir))
	return saveHistory(hist)
}

func cmdRemove(dir string, force bool) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	aliases, err := loadAliases()
	if err != nil {
		return err
	}

	removeKey, locations, err := resolveRemoveEntry(hist, aliases, dir)
	if err != nil {
		return err
	}
	if locations == nil {
		absDir, err := filepath.Abs(expandConfiguredPath(dir))
		if err != nil {
			return fmt.Errorf("resolve: %w", err)
		}
		fmt.Printf("hint: no recorded entry for %s\n", displayPath(absDir))
		return nil
	}

	locations, ok := hist[removeKey]
	if !ok {
		fmt.Printf("hint: no recorded entry for %s\n", displayPath(removeKey))
		return nil
	}

	if len(locations) > 1 {
		if !force {
			return fmt.Errorf("%s has movement history:\n  use `mvd --rm -f %s`\n  to clear it", displayPath(removeKey), removeKey)
		}
		fmt.Printf("hint: removing %s will clear %d history entries\n", displayPath(removeKey), len(locations)-1)
	}

	delete(hist, removeKey)
	fmt.Printf("removed: %s\n", displayPath(removeKey))
	return saveHistory(hist)
}

func cmdRebase(dir, newDir string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	origKey, locations, absNewDir, err := resolveRebaseEntries(hist, dir, newDir)
	if err != nil {
		return err
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
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	absSrc, origKey, locations, err := resolveClearEntry(hist, src)
	if err != nil {
		return err
	}
	if locations == nil {
		fmt.Printf("no movement history for %s\n", displayPath(absSrc))
		return nil
	}

	delete(hist, origKey)
	fmt.Printf("cleared history for %s (%d entries)\n", displayPath(absSrc), len(locations))
	return saveHistory(hist)
}

func cmdWhich(src string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	aliases, err := loadAliases()
	if err != nil {
		return err
	}

	matches, err := findWhichMatches(hist, aliases, src)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("%s not found", src)
	}
	for _, match := range matches {
		fmt.Printf("%s (%s)\n", displayPath(match.path), match.label)
	}
	return nil
}

func cmdListAll() error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	aliases, err := loadAliases()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		fmt.Println("(no mv history)")
		return nil
	}
	aliasesByRoot := groupAliasesByRoot(aliases)

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
		fmt.Printf("%s%s\n", displayPath(key), formatAliases(aliasesByRoot[key]))
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

func groupAliasesByRoot(aliases map[string]string) map[string][]string {
	aliasesByRoot := make(map[string][]string)
	for alias, root := range aliases {
		aliasesByRoot[root] = append(aliasesByRoot[root], alias)
	}
	for root := range aliasesByRoot {
		sort.Strings(aliasesByRoot[root])
	}
	return aliasesByRoot
}

func formatAliases(aliases []string) string {
	if len(aliases) == 0 {
		return ""
	}
	return fmt.Sprintf(" (aliases: %s)", strings.Join(aliases, ", "))
}

func cmdPrint(src string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	fullPath, err := resolvePrintPath(hist, src)
	if err != nil {
		return err
	}

	fmt.Printf("%s -> %s\n", displayPath(fullPath), fullPath)
	return nil
}

func cmdList(src string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	absSrc, locations, err := resolveListEntry(hist, src)
	if err != nil {
		return err
	}
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

func expandConfiguredPath(path string) string {
	cfg := loadDisplayConfig()
	if cfg == nil {
		return path
	}
	return llsconfig.ExpandPath(path, cfg.Envs)
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

func aliasesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mvd", "aliases.json")
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

func loadAliases() (map[string]string, error) {
	p := aliasesPath()
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read aliases: %w", err)
	}
	var aliases map[string]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, fmt.Errorf("parse aliases: %w", err)
	}
	if aliases == nil {
		aliases = make(map[string]string)
	}
	return aliases, nil
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

func saveAliases(aliases map[string]string) error {
	p := aliasesPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal aliases: %w", err)
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("write aliases: %w", err)
	}
	return nil
}
