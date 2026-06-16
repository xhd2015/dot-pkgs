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

// MoveEntry records an explicit from→to transition with endpoint types.
// from_type / to_type are "main" (repo directory) or "worktree".
type MoveEntry struct {
	From     string `json:"from"`
	FromType string `json:"from_type"`
	To       string `json:"to"`
	ToType   string `json:"to_type"`
	Branch   string `json:"branch,omitempty"`
}

type ProjectEntry struct {
	Root      string          `json:"root,omitempty"`
	Locations []LocationEntry `json:"locations,omitempty"`
	Moves     []MoveEntry     `json:"moves,omitempty"`
	Aliases   []string        `json:"aliases,omitempty"`
}

type HistoryFile struct {
	Version  string                  `json:"version"`
	Projects map[string]ProjectEntry `json:"projects"`
}

type History map[string][]LocationEntry