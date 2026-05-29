package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func cmdMove(src, dst string) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	aliases, err := loadAliases()
	if err != nil {
		return err
	}

	origKey, locations, absSrc, err := resolveMoveSource(hist, aliases, src)
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
