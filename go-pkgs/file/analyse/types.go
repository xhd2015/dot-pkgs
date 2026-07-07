package analyse

// EntryKind classifies a top-level HOME child.
type EntryKind string

const (
	EntryKindFile EntryKind = "file"
	EntryKindDir  EntryKind = "dir"
)

// ChildLine is one immediate child with deep-aggregated size.
type ChildLine struct {
	Name      string `json:"name"`
	Bytes     int64  `json:"bytes"`
	SizeHuman string `json:"size_human"`
}

// SemanticLine is one enricher field for a tool directory.
type SemanticLine struct {
	Key       string `json:"key"`
	Count     string `json:"count,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Bytes     int64  `json:"bytes"`
	SizeHuman string `json:"size_human"`
	Extra     string `json:"extra,omitempty"`
}

// Aggregates holds per-entry rollups shown after semantic lines.
type Aggregates struct {
	GitRepos        int `json:"git_repos"`
	LinkedWorktrees int `json:"linked_worktrees"`
	NodeModulesDirs int `json:"node_modules_dirs"`
}

// EntryResult is one scanned top-level HOME child.
type EntryResult struct {
	Name       string         `json:"name"`
	Kind       EntryKind      `json:"kind"`
	Bytes      int64          `json:"bytes"`
	SizeHuman  string         `json:"size_human,omitempty"`
	Lines      string         `json:"lines,omitempty"`
	Children   []ChildLine    `json:"children,omitempty"`
	Semantic   []SemanticLine `json:"semantic,omitempty"`
	Aggregates Aggregates     `json:"aggregates"`
}

// ScanSummary holds global rollups for the summary block and done frame.
type ScanSummary struct {
	Home              string `json:"home"`
	EntryCount        int    `json:"entries"`
	DirCount          int    `json:"dirs"`
	FileCount         int    `json:"files"`
	TotalBytes        int64  `json:"total_bytes"`
	TotalHuman        string `json:"total_human"`
	GitRepos          int    `json:"git_repos"`
	LinkedWorktrees   int    `json:"linked_worktrees"`
	NodeModulesDirs   int    `json:"node_modules_dirs"`
	CodexSessions     int    `json:"codex_sessions,omitempty"`
	CodexSkills       int    `json:"codex_skills,omitempty"`
	GrokSessions      int    `json:"grok_sessions,omitempty"`
	GrokProjects      int    `json:"grok_projects,omitempty"`
	GrokSkills        int    `json:"grok_skills,omitempty"`
	CursorProjects    int    `json:"cursor_projects,omitempty"`
	CursorChats       int    `json:"cursor_chats,omitempty"`
	KHKnowledges      int    `json:"knowledge_hub_knowledges,omitempty"`
	KIAgents          int    `json:"knowledge_index_agents,omitempty"`
	OpenclawAgents    int    `json:"openclaw_agents,omitempty"`
	HasCodex          bool   `json:"has_codex"`
	HasGrok           bool   `json:"has_grok"`
	HasCursor         bool   `json:"has_cursor"`
	HasKnowledgeHub   bool   `json:"has_knowledge_hub"`
	HasKnowledgeIndex bool   `json:"has_knowledge_index"`
	HasOpenclaw       bool   `json:"has_openclaw"`
	Largest           []LargestEntry `json:"largest_entries"`
}

// LargestEntry is one top-N entry by size.
type LargestEntry struct {
	Name      string `json:"name"`
	Bytes     int64  `json:"bytes"`
	SizeHuman string `json:"size_human"`
}

// Options configures a HOME scan.
type Options struct {
	Home string
	// OnEntry is called after each top-level HOME child is fully scanned.
	// Return non-nil error to abort scan (propagated from Scan).
	OnEntry func(EntryResult) error
}