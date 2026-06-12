package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func Load(path string) (History, map[string]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(History), make(map[string]string), nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read history: %w", err)
	}

	var file HistoryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, nil, fmt.Errorf("parse history: %w", err)
	}

	aliases := extractAliases(file.Projects)

	if file.Projects != nil {
		if hasV2Format(file.Projects) {
			hist := make(History, len(file.Projects))
			for key, proj := range file.Projects {
				hist[key] = LocationsFromMoves(proj.Root, proj.Moves)
			}
			return mergeChains(hist), aliases, nil
		}
		hist := make(History, len(file.Projects))
		for key, proj := range file.Projects {
			hist[key] = proj.Locations
		}
		return mergeChains(hist), aliases, nil
	}

	var legacy map[string][]string
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, nil, fmt.Errorf("parse history: %w", err)
	}
	hist := make(History, len(legacy))
	for key, locs := range legacy {
		entries := make([]LocationEntry, len(locs))
		for i, loc := range locs {
			entries[i] = LocationEntry{Path: loc}
		}
		hist[key] = entries
	}
	return mergeChains(hist), aliases, nil
}

func extractAliases(projects map[string]ProjectEntry) map[string]string {
	aliases := make(map[string]string)
	if projects == nil {
		return aliases
	}
	for key, proj := range projects {
		for _, alias := range proj.Aliases {
			aliases[alias] = key
		}
	}
	return aliases
}

func Save(path string, hist History, aliases map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	aliasesByRoot := make(map[string][]string)
	for alias, root := range aliases {
		aliasesByRoot[root] = append(aliasesByRoot[root], alias)
	}
	for root := range aliasesByRoot {
		sort.Strings(aliasesByRoot[root])
	}

	projects := make(map[string]ProjectEntry, len(hist))
	for key, locs := range hist {
		root, moves := DeriveMoves(locs)
		proj := ProjectEntry{Root: root, Moves: moves}
		if aliasList, ok := aliasesByRoot[key]; ok {
			proj.Aliases = aliasList
		}
		projects[key] = proj
	}
	file := HistoryFile{
		Version:  "2.0",
		Projects: projects,
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

func hasV2Format(projects map[string]ProjectEntry) bool {
	for _, proj := range projects {
		if proj.Root != "" {
			return true
		}
	}
	return false
}

func mergeChains(hist History) History {
	for {
		merged := false
		for key, locs := range hist {
			if len(locs) == 0 {
				continue
			}
			tail := locs[len(locs)-1].Path
			if tail == key {
				continue
			}
			next, ok := hist[tail]
			if !ok || len(next) == 0 {
				continue
			}
			combined := append([]LocationEntry{}, locs...)
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
