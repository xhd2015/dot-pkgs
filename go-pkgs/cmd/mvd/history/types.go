package history

type GitInfo struct {
	Type     string `json:"type"`
	MainRepo string `json:"main_repo,omitempty"`
	Branch   string `json:"branch,omitempty"`
}

type LocationEntry struct {
	Path string   `json:"path"`
	Git  *GitInfo `json:"git,omitempty"`
}

type MoveEntry struct {
	Prev    string `json:"prev"`
	Current string `json:"current"`
	Type    string `json:"type"`
	Branch  string `json:"branch,omitempty"`
}

type ProjectEntry struct {
	Root      string          `json:"root,omitempty"`
	Locations []LocationEntry `json:"locations,omitempty"`
	Moves     []MoveEntry     `json:"moves,omitempty"`
}

type HistoryFile struct {
	Version  string                  `json:"version"`
	Projects map[string]ProjectEntry `json:"projects"`
}

type History map[string][]LocationEntry
