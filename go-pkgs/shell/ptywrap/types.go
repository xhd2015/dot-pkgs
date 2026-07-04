package ptywrap

import "time"

// SessionInfo is the JSON representation of a session returned to clients.
type SessionInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Command   []string  `json:"command,omitempty"`
	Cwd       string    `json:"cwd"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Connected bool      `json:"connected"`

	ObserverCount   int  `json:"observer_count"`
	AttacherCount   int  `json:"attacher_count"`
	WriterConnected bool `json:"writer_connected"`
}

// SessionsResponse holds paginated terminal sessions response.
type SessionsResponse struct {
	Sessions   []SessionInfo `json:"sessions"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
}

// CreateRequest is the REST body for creating a session.
type CreateRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Cwd     string   `json:"cwd"`
	Name    string   `json:"name"`
}

// RenameRequest is the REST body for renaming a session.
type RenameRequest struct {
	Name string `json:"name"`
}

// ControlMessage is a JSON message sent from client to control the terminal.
type ControlMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

// SpawnOptions configures how a PTY session is started.
type SpawnOptions struct {
	Shell      string
	ShellFlags []string
	ExtraPaths []string
	PS1        string
}