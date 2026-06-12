package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Load(path string) (History, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(History), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}

	var file HistoryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}

	if file.Projects != nil {
		if hasV2Format(file.Projects) {
			hist := make(History, len(file.Projects))
			for key, proj := range file.Projects {
				hist[key] = LocationsFromMoves(proj.Root, proj.Moves)
			}
			return mergeChains(hist), nil
		}
		hist := make(History, len(file.Projects))
		for key, proj := range file.Projects {
			hist[key] = proj.Locations
		}
		return mergeChains(hist), nil
	}

	var legacy map[string][]string
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	hist := make(History, len(legacy))
	for key, locs := range legacy {
		entries := make([]LocationEntry, len(locs))
		for i, loc := range locs {
			entries[i] = LocationEntry{Path: loc}
		}
		hist[key] = entries
	}
	return mergeChains(hist), nil
}

func Save(path string, hist History) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	projects := make(map[string]ProjectEntry, len(hist))
	for key, locs := range hist {
		root, moves := DeriveMoves(locs)
		projects[key] = ProjectEntry{Root: root, Moves: moves}
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
