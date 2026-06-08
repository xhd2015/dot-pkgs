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

	absProject, err := resolveInputPath(project)
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

func resolveInputPath(path string) (string, error) {
	absPath, err := filepath.Abs(expandConfiguredPath(path))
	if err != nil {
		return "", err
	}
	return absPath, nil
}

func resolveExistingPath(path string) (string, error) {
	expanded := expandConfiguredPath(path)
	if _, err := os.Stat(expanded); err != nil {
		return "", fmt.Errorf("%s does not exist", displayPath(expanded))
	}
	return filepath.Abs(expanded)
}

func resolveAddPath(dir string) (string, error) {
	absDir, err := resolveExistingPath(dir)
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}
	if info, err := os.Stat(absDir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s does not exist", displayPath(absDir))
	}
	return absDir, nil
}

func isRecordedPath(hist History, path string) bool {
	_, locations := findEntry(hist, path)
	return locations != nil
}

func resolveRemoveEntry(hist History, aliases map[string]string, dir string) (string, []LocationEntry, error) {
	if k, locs, ok, err := resolveBasename(hist, dir); ok {
		return k, locs, nil
	} else if err != nil {
		return "", nil, err
	}

	if isBareBaseName(dir) {
		if _, err := os.Lstat(dir); os.IsNotExist(err) {
			origKey, locations, err := findEntryByAlias(hist, aliases, dir)
			if err != nil {
				return "", nil, err
			}
			if locations != nil {
				return origKey, locations, nil
			}
		}
	}

	absDir, err := resolveInputPath(dir)
	if err != nil {
		return "", nil, fmt.Errorf("resolve: %w", err)
	}
	if locations, ok := hist[absDir]; ok {
		return absDir, locations, nil
	}
	return "", nil, nil
}

func resolveRebaseEntries(hist History, dir, newDir string) (string, []LocationEntry, string, error) {
	if k, locs, ok, err := resolveBasename(hist, dir); ok {
		absNewDir, err := filepath.Abs(newDir)
		if err != nil {
			return "", nil, "", fmt.Errorf("resolve new-dir: %w", err)
		}
		if otherKey, otherLocations := findEntry(hist, absNewDir); otherLocations != nil && otherKey != k {
			return "", nil, "", fmt.Errorf("%s is already recorded in another entry", absNewDir)
		}
		return k, locs, absNewDir, nil
	} else if err != nil {
		return "", nil, "", err
	}

	absDir, err := resolveInputPath(dir)
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
	if k, locs, ok, err := resolveBasename(hist, src); ok {
		return k, k, locs, nil
	} else if err != nil {
		return "", "", nil, err
	}

	absSrc, err := resolveInputPath(src)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve: %w", err)
	}

	origKey, locations := findEntry(hist, absSrc)
	return absSrc, origKey, locations, nil
}

func resolveMoveSource(hist History, aliases map[string]string, src string) (string, []LocationEntry, string, error) {
	if k, locs, ok, err := resolveBasename(hist, src); ok {
		if len(locs) == 0 {
			return "", nil, "", fmt.Errorf("empty mv history for %s", src)
		}
		return k, locs, locs[len(locs)-1].Path, nil
	} else if err != nil {
		return "", nil, "", err
	}

	if isBareBaseName(src) {
		if _, err := os.Lstat(src); os.IsNotExist(err) {
			origKey, locations, err := findEntryByAlias(hist, aliases, src)
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
	}

	absSrc, err := resolveInputPath(src)
	if err != nil {
		return "", nil, "", fmt.Errorf("resolve src: %w", err)
	}

	origKey, locations := findEntry(hist, absSrc)
	if locations == nil {
		if _, err := os.Stat(absSrc); err != nil {
			return "", nil, "", fmt.Errorf("%s does not exist", displayPath(absSrc))
		}
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
	if k, locs, ok, err := resolveBasename(hist, src); ok {
		return k, locs, nil
	} else if err != nil {
		return "", nil, err
	}

	absSrc, err := resolveInputPath(src)
	if err != nil {
		return "", nil, fmt.Errorf("resolve: %w", err)
	}
	return absSrc, findLocations(hist, absSrc), nil
}

func resolvePrintPath(hist History, src string) (string, error) {
	if _, locs, ok, err := resolveBasename(hist, src); ok {
		if len(locs) == 0 {
			return "", fmt.Errorf("empty mv history for %s", src)
		}
		return locs[0].Path, nil
	} else if err != nil {
		return "", err
	}

	absSrc, err := resolveInputPath(src)
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}
	return absSrc, nil
}

func resolveBackEntry(hist History, src string) (string, []LocationEntry, error) {
	if k, locs, ok, err := resolveBasename(hist, src); ok {
		return k, locs, nil
	} else if err != nil {
		return "", nil, err
	}

	absSrc, err := resolveInputPath(src)
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

	absSrc, err := resolveInputPath(src)
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

func resolveBasename(hist History, name string) (string, []LocationEntry, bool, error) {
	if !isBareBaseName(name) {
		return "", nil, false, nil
	}
	if _, err := os.Lstat(name); !os.IsNotExist(err) {
		return "", nil, false, nil
	}
	k, l, err := findEntryByRootBaseName(hist, name)
	if err != nil {
		return "", nil, false, err
	}
	if l == nil {
		return "", nil, false, nil
	}
	return k, l, true, nil
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
