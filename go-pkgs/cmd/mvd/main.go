package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/xhd2015/less-gen/flags"
)

// History tracks movement history for each file.
// Key: original absolute path, Value: list of locations (oldest first).
type History map[string][]string

const help = `
Usage: mvd [OPTIONS] SRC [DST]

Move a file/directory and track its location history.

Commands:
  mvd SRC DST          Move SRC to DST (tracked)
  mvd --list [SRC]     Show location history (all if SRC omitted)
  mvd --back SRC       Move SRC back to its previous location
  mvd --clear SRC      Clear movement history for SRC

Options:
  --list               Show location history
  --back               Move back to previous location
  --clear              Clear movement history for a specific file
  -h, --help           Show this help message
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "mvd: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var list, back, clear bool
	args, err := flags.Bool("--list", &list).
		Bool("--back", &back).
		Bool("--clear", &clear).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
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
		fmt.Printf("no movement history for %s\n", absSrc)
		return nil
	}

	delete(hist, origKey)
	fmt.Printf("cleared history for %s (%d entries)\n", absSrc, len(locations))
	return saveHistory(hist)
}

func cmdMove(src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("resolve src: %w", err)
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

	hist, err := loadHistory()
	if err != nil {
		return err
	}

	locations := hist[absSrc]
	if len(locations) == 0 {
		locations = []string{absSrc, absDst}
	} else {
		locations = append(locations, absDst)
	}

	origKey := locations[0]
	delete(hist, absSrc)
	hist[origKey] = locations

	return saveHistory(hist)
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
		fmt.Printf("%s\n", key)
		for j, loc := range locations {
			marker := "  "
			if loc == current {
				marker = "* "
			}
			if j == 0 {
				fmt.Printf("  %s%s  (original)\n", marker, loc)
			} else {
				fmt.Printf("  %s%s\n", marker, loc)
			}
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
			fmt.Printf("%s%s  (original)\n", marker, loc)
		} else {
			fmt.Printf("%s%s\n", marker, loc)
		}
	}
	return nil
}

func cmdBack(src string) error {
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
		return fmt.Errorf("no mv history for %s", absSrc)
	}

	last := locations[len(locations)-1]
	if last != absSrc {
		return fmt.Errorf("current position mismatch: expected %s at end of history, got %s", absSrc, last)
	}

	if len(locations) <= 1 {
		return fmt.Errorf("already at original position: %s", absSrc)
	}

	prev := locations[len(locations)-2]
	if err := os.Rename(absSrc, prev); err != nil {
		return fmt.Errorf("move back: %w", err)
	}
	fmt.Printf("moved back: %s → %s\n", absSrc, prev)

	locations = locations[:len(locations)-1]
	if len(locations) <= 1 {
		delete(hist, origKey)
	} else {
		hist[origKey] = locations
	}

	return saveHistory(hist)
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
	return hist, nil
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
