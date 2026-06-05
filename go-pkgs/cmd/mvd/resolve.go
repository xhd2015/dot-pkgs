package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type whichMatch struct {
	path  string
	label string
}

func resolveAliasProject(hist History, project string) (string, error) {
	if isBareBaseName(project) {
		_, locations, err := findEntryByRootBaseName(hist, project)
		if err != nil {
			return "", err
		}
		if locations != nil {
			if len(locations) == 0 {
				return "", fmt.Errorf("empty mv history for %s", project)
			}
			return locations[0].Path, nil
		}
	}

	absProject, err := filepath.Abs(expandConfiguredPath(project))
	if err != nil {
		return "", fmt.Errorf("resolve project: %w", err)
	}
	_, locations := findEntry(hist, absProject)
	if locations == nil {
		return "", fmt.Errorf("no recorded project for %s", absProject)
	}
	if len(locations) == 0 {
		return "", fmt.Errorf("empty mv history for %s", absProject)
	}
	return locations[0].Path, nil
}

func validateAliasName(alias string) error {
	if !isBareBaseName(alias) {
		return fmt.Errorf("alias must be a bare name: %s", alias)
	}
	return nil
}

func resolveAddPath(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}
	return absDir, nil
}

func isRecordedPath(hist History, path string) bool {
	_, locations := findEntry(hist, path)
	return locations != nil
}

func resolveRemoveEntry(hist History, aliases map[string]string, dir string) (string, []LocationEntry, error) {
	if useRootBaseNameShortcut(dir) {
		origKey, locations, err := findEntryByRootBaseName(hist, dir)
		if err != nil {
			return "", nil, err
		}
		if locations != nil {
			return origKey, locations, nil
		}
		origKey, locations, err = findEntryByAlias(hist, aliases, dir)
		if err != nil {
			return "", nil, err
		}
		if locations != nil {
			return origKey, locations, nil
		}
	}

	absDir, err := filepath.Abs(expandConfiguredPath(dir))
	if err != nil {
		return "", nil, fmt.Errorf("resolve: %w", err)
	}
	if locations, ok := hist[absDir]; ok {
		return absDir, locations, nil
	}
	return "", nil, nil
}

func resolveRebaseEntries(hist History, dir, newDir string) (string, []LocationEntry, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, "", fmt.Errorf("resolve dir: %w", err)
	}
	absNewDir, err := filepath.Abs(newDir)
	if err != nil {
		return "", nil, "", fmt.Errorf("resolve new-dir: %w", err)
	}

	origKey, locations := findEntry(hist, absDir)
	if locations == nil {
		return "", nil, "", fmt.Errorf("no recorded entry for %s", absDir)
	}

	if otherKey, otherLocations := findEntry(hist, absNewDir); otherLocations != nil && otherKey != origKey {
		return "", nil, "", fmt.Errorf("%s is already recorded in another entry", absNewDir)
	}

	return origKey, locations, absNewDir, nil
}

func resolveClearEntry(hist History, src string) (string, string, []LocationEntry, error) {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve: %w", err)
	}

	origKey, locations := findEntry(hist, absSrc)
	return absSrc, origKey, locations, nil
}

func resolveMoveSource(hist History, aliases map[string]string, src string) (string, []LocationEntry, string, error) {
	if useRootBaseNameShortcut(src) {
		origKey, locations, err := findEntryByRootBaseName(hist, src)
		if err != nil {
			return "", nil, "", err
		}
		if locations != nil {
			if len(locations) == 0 {
				return "", nil, "", fmt.Errorf("empty mv history for %s", src)
			}
			return origKey, locations, locations[len(locations)-1].Path, nil
		}
		origKey, locations, err = findEntryByAlias(hist, aliases, src)
		if err != nil {
			return "", nil, "", err
		}
		if locations != nil {
			if len(locations) == 0 {
				return "", nil, "", fmt.Errorf("empty mv history for alias %s", src)
			}
			return origKey, locations, locations[len(locations)-1].Path, nil
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

	last := locations[len(locations)-1].Path
	root := locations[0].Path
	if absSrc != root && absSrc != last {
		return "", nil, "", fmt.Errorf("current position mismatch: expected %s at end of history, got %s", absSrc, last)
	}

	return origKey, locations, last, nil
}

func resolveListEntry(hist History, src string) (string, []LocationEntry, error) {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return "", nil, fmt.Errorf("resolve: %w", err)
	}
	return absSrc, findLocations(hist, absSrc), nil
}

func resolvePrintPath(hist History, src string) (string, error) {
	if useRootBaseNameShortcut(src) {
		_, locations, err := findEntryByRootBaseName(hist, src)
		if err != nil {
			return "", err
		}
		if locations != nil {
			if len(locations) == 0 {
				return "", fmt.Errorf("empty mv history for %s", src)
			}
			return locations[0].Path, nil
		}
	}

	absSrc, err := filepath.Abs(expandConfiguredPath(src))
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}
	return absSrc, nil
}

func resolveBackEntry(hist History, src string) (string, []LocationEntry, error) {
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

	last := locations[len(locations)-1].Path
	root := locations[0].Path
	if absSrc != root && absSrc != last {
		return "", nil, fmt.Errorf("current position mismatch: expected %s at end of history, got %s", absSrc, last)
	}

	return origKey, locations, nil
}

func findEntryByAlias(hist History, aliases map[string]string, alias string) (string, []LocationEntry, error) {
	root, ok := aliases[alias]
	if !ok {
		return "", nil, nil
	}
	origKey, locations := findEntry(hist, root)
	if locations == nil {
		return "", nil, fmt.Errorf("alias %s points to unrecorded project %s", alias, root)
	}
	return origKey, locations, nil
}

func findWhichMatches(hist History, aliases map[string]string, src string) ([]whichMatch, error) {
	var matches []whichMatch

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return nil, fmt.Errorf("resolve src: %w", err)
	}
	if _, err := os.Lstat(absSrc); err == nil {
		matches = append(matches, whichMatch{path: absSrc, label: "local"})
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat src: %w", err)
	}

	if !isBareBaseName(src) {
		if _, locations := findEntry(hist, absSrc); locations != nil {
			if len(locations) == 0 {
				return nil, fmt.Errorf("empty mv history for %s", absSrc)
			}
			matches = append(matches, whichMatch{path: locations[len(locations)-1].Path, label: "tracked"})
		}
		return matches, nil
	}

	for _, locations := range findEntriesByRootBaseName(hist, src) {
		if len(locations) == 0 {
			return nil, fmt.Errorf("empty mv history for %s", src)
		}
		matches = append(matches, whichMatch{path: locations[len(locations)-1].Path, label: "project basename"})
	}

	if root, ok := aliases[src]; ok {
		_, locations := findEntry(hist, root)
		if locations == nil {
			return nil, fmt.Errorf("alias %s points to unrecorded project %s", src, root)
		}
		if len(locations) == 0 {
			return nil, fmt.Errorf("empty mv history for alias %s", src)
		}
		matches = append(matches, whichMatch{path: locations[len(locations)-1].Path, label: "alias: " + src})
	}

	return matches, nil
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

func findEntryByRootBaseName(hist History, baseName string) (string, []LocationEntry, error) {
	var matchedKey string
	var matchedLocations []LocationEntry
	var matchedRoots []string

	for _, match := range findEntriesByRootBaseName(hist, baseName) {
		root := match[0].Path
		matchedRoots = append(matchedRoots, root)
		matchedKey = root
		matchedLocations = match
	}

	if len(matchedRoots) > 1 {
		sort.Strings(matchedRoots)
		return "", nil, fmt.Errorf("ambiguous root basename %s matches multiple roots: %v", baseName, displayPaths(matchedRoots))
	}
	return matchedKey, matchedLocations, nil
}

func findEntriesByRootBaseName(hist History, baseName string) [][]LocationEntry {
	roots := make([]string, 0, len(hist))
	byRoot := make(map[string][]LocationEntry)
	for _, locations := range hist {
		if len(locations) == 0 {
			continue
		}
		root := locations[0].Path
		if filepath.Base(root) != baseName {
			continue
		}
		roots = append(roots, root)
		byRoot[root] = locations
	}
	sort.Strings(roots)

	matches := make([][]LocationEntry, 0, len(roots))
	for _, root := range roots {
		matches = append(matches, byRoot[root])
	}
	return matches
}

func findLocations(hist History, path string) []LocationEntry {
	_, locs := findEntry(hist, path)
	return locs
}

func findEntry(hist History, path string) (string, []LocationEntry) {
	if locs, ok := hist[path]; ok {
		return path, locs
	}
	for key, locs := range hist {
		for _, loc := range locs {
			if loc.Path == path {
				return key, locs
			}
		}
		_ = key
	}
	return "", nil
}
