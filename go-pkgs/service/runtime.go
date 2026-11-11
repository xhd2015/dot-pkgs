package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// RuntimeStateVersion is bumped when the on-disk record shape changes. Readers
// ignore a file whose version they do not know instead of misreading pids.
const RuntimeStateVersion = 1

// RuntimeRecord is the pid a host left behind for one service. It exists so the
// next host process can tell "a service I started earlier is still running"
// from "that pid now belongs to something else".
type RuntimeRecord struct {
	PID       int    `json:"pid"`
	PGID      int    `json:"pgid,omitempty"`
	StartedAt string `json:"startedAt,omitempty"`
	// Command is the definition's command at start time, used to detect pid
	// recycling before the record is trusted.
	Command string `json:"command,omitempty"`
	// Detached records how the process was started, not what the definition
	// says now. Editing a definition must not change the lifetime of a process
	// that is already running, and the next host instance needs the start-time
	// value to decide between reclaiming and adopting it.
	Detached bool `json:"detached,omitempty"`
}

// RuntimeState is the whole runtime file: the pids one host instance recorded.
type RuntimeState struct {
	Version   int                      `json:"version"`
	Processes map[string]RuntimeRecord `json:"processes,omitempty"`
}

// RuntimeStore reads and writes the runtime pid file.
//
// Writes are atomic (temp file + rename): a host killed mid-write must not
// leave a truncated file that loses every running pid.
type RuntimeStore struct {
	mu   sync.Mutex
	path string
}

// NewRuntimeStore returns a store backed by path. An empty path makes Load
// always empty and Save a no-op, which is how a host with no data directory
// disables orphan tracking.
func NewRuntimeStore(path string) *RuntimeStore {
	return &RuntimeStore{path: path}
}

// Path returns the file backing the store.
func (s *RuntimeStore) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// emptyRuntimeState is a state with no records. The map is non-nil so callers
// can record into a freshly loaded state without a nil-map panic.
func emptyRuntimeState() RuntimeState {
	return RuntimeState{Version: RuntimeStateVersion, Processes: map[string]RuntimeRecord{}}
}

// Load returns the recorded state. A missing, unreadable, corrupt, or
// unknown-version file yields an empty state: losing orphan records only costs
// a leaked process, while refusing to start would cost the whole feature.
func (s *RuntimeStore) Load() RuntimeState {
	empty := emptyRuntimeState()
	if s == nil || s.path == "" {
		return empty
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *RuntimeStore) loadLocked() RuntimeState {
	empty := emptyRuntimeState()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return empty
	}
	var state RuntimeState
	if err := json.Unmarshal(data, &state); err != nil {
		return empty
	}
	if state.Version != RuntimeStateVersion {
		return empty
	}
	if state.Processes == nil {
		state.Processes = map[string]RuntimeRecord{}
	}
	state.Version = RuntimeStateVersion
	return state
}

// Save replaces the whole state atomically.
func (s *RuntimeStore) Save(state RuntimeState) error {
	if s == nil || s.path == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(state)
}

func (s *RuntimeStore) saveLocked(state RuntimeState) error {
	state.Version = RuntimeStateVersion
	if state.Processes == nil {
		state.Processes = map[string]RuntimeRecord{}
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create runtime state directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Record remembers that id is running as pid.
func (s *RuntimeStore) Record(id string, rec RuntimeRecord) error {
	if s == nil || s.path == "" || id == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.loadLocked()
	state.Processes[id] = rec
	return s.saveLocked(state)
}

// Forget drops the record for id. It is a no-op when the record is already
// gone, so every exit path can call it unconditionally.
func (s *RuntimeStore) Forget(id string) error {
	if s == nil || s.path == "" || id == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.loadLocked()
	if _, ok := state.Processes[id]; !ok {
		return nil
	}
	delete(state.Processes, id)
	return s.saveLocked(state)
}

// IDs returns the recorded service ids, sorted.
func (s RuntimeState) IDs() []string {
	ids := make([]string, 0, len(s.Processes))
	for id := range s.Processes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ParseRecordTime parses a record's StartedAt, returning the zero time when it
// is absent or malformed.
func ParseRecordTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}
