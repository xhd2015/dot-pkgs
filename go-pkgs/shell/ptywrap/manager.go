package ptywrap

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Manager tracks PTY sessions in memory.
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*session
	counter  int
	Spawn    SpawnOptions
}

// NewManager creates an empty session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*session),
	}
}

// DefaultManager is the process-wide session manager used by RegisterAPI.
var DefaultManager = NewManager()

func (m *Manager) createShell(name, cwd string) (*session, error) {
	cmd, ptmx, command, err := startDefaultShellPTY(cwd, m.Spawn)
	if err != nil {
		return nil, err
	}
	return m.registerSession(name, cwd, command, cmd, ptmx)
}

func (m *Manager) createCommand(name, cwd string, command []string) (*session, error) {
	cmd, ptmx, resolved, err := startPTY(command, cwd, m.Spawn)
	if err != nil {
		return nil, err
	}
	return m.registerSession(name, cwd, resolved, cmd, ptmx)
}

// CreateCommand starts a new PTY session running command in cwd.
func (m *Manager) CreateCommand(name, cwd string, command []string) (SessionInfo, error) {
	s, err := m.createCommand(name, cwd, command)
	if err != nil {
		return SessionInfo{}, err
	}
	return s.info(false), nil
}

func (m *Manager) registerSession(name, cwd string, command []string, cmd *exec.Cmd, ptmx *os.File) (*session, error) {
	m.mu.Lock()
	m.counter++
	id := fmt.Sprintf("session-%d", m.counter)
	m.mu.Unlock()

	if name == "" {
		name = "Terminal"
	}
	if cwd == "" {
		cwd = cmd.Dir
	}

	s := &session{
		id:        id,
		name:      name,
		command:   append([]string(nil), command...),
		cwd:       cwd,
		createdAt: time.Now(),
		cmd:       cmd,
		ptmx:      ptmx,
		cols:      80,
		rows:      24,
		done:      make(chan struct{}),
	}

	m.mu.Lock()
	if m.sessions == nil {
		m.sessions = make(map[string]*session)
	}
	m.sessions[id] = s
	m.mu.Unlock()

	go s.readLoop()
	return s, nil
}

func (m *Manager) get(id string) *session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[id]
}

// Scrollback returns a readonly copy of the session scrollback buffer.
func (m *Manager) Scrollback(id string) []byte {
	s := m.get(id)
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.scrollback))
	copy(out, s.scrollback)
	return out
}

// WriteInput writes bytes to the session PTY master (e.g. prompt injection).
func (m *Manager) WriteInput(id string, data []byte) error {
	s := m.get(id)
	if s == nil {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.exited {
		return fmt.Errorf("session exited: %s", id)
	}
	_, err := s.ptmx.Write(data)
	return err
}

// Wait blocks until the session process exits.
func (m *Manager) Wait(id string) error {
	s := m.get(id)
	if s == nil {
		return fmt.Errorf("session not found: %s", id)
	}
	<-s.done
	if s.cmd != nil && s.cmd.ProcessState != nil && !s.cmd.ProcessState.Success() {
		return fmt.Errorf("session exited with code %d", s.cmd.ProcessState.ExitCode())
	}
	return nil
}



func (m *Manager) rename(id, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	s.name = name
	return nil
}

func (m *Manager) listPaginated(page, pageSize int) *SessionsResponse {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionList := make([]*session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessionList = append(sessionList, s)
	}

	sort.Slice(sessionList, func(i, j int) bool {
		return sessionList[i].createdAt.Before(sessionList[j].createdAt)
	})

	total := len(sessionList)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedSessions []*session
	if start < total {
		pagedSessions = sessionList[start:end]
	}

	sessions := make([]SessionInfo, 0, len(pagedSessions))
	for _, s := range pagedSessions {
		s.mu.Lock()
		connected := s.conn != nil
		s.mu.Unlock()
		sessions = append(sessions, s.info(connected))
	}

	return &SessionsResponse{
		Sessions:   sessions,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// RegisterExternal adds a session backed by an existing PTY process (e.g. SSH).
func (m *Manager) RegisterExternal(name, cwd string, command []string, cmd *exec.Cmd, ptmx *os.File) (string, error) {
	s, err := m.registerSession(name, cwd, command, cmd, ptmx)
	if err != nil {
		return "", err
	}
	return s.id, nil
}

// Remove stops and deletes a session by id.
func (m *Manager) Remove(id string) {
	m.remove(id)
}

func (m *Manager) remove(id string) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()

	if !ok {
		return
	}
	s.close()
}

// RegisterTestSessions seeds sessions for doctest harnesses.
func RegisterTestSessions(m *Manager, infos []SessionInfo) error {
	for _, info := range infos {
		command := info.Command
		if len(command) == 0 {
			command = []string{"sleep", "3600"}
		}
		name := info.Name
		if name == "" {
			name = "test"
		}
		s, err := m.createCommand(name, info.Cwd, command)
		if err != nil {
			return err
		}
		if info.ID != "" && s.id != info.ID {
			m.mu.Lock()
			delete(m.sessions, s.id)
			s.id = info.ID
			m.sessions[info.ID] = s
			if n, ok := parseSessionCounter(info.ID); ok && n > m.counter {
				m.counter = n
			}
			m.mu.Unlock()
		}
	}
	return nil
}

func parseSessionCounter(id string) (int, bool) {
	const prefix = "session-"
	if !strings.HasPrefix(id, prefix) {
		return 0, false
	}
	n, err := strconv.Atoi(id[len(prefix):])
	if err != nil {
		return 0, false
	}
	return n, true
}