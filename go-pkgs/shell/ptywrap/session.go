package ptywrap

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

type attachRole string

const (
	roleWriter   attachRole = "writer"
	roleObserver attachRole = "observer"
	roleSnapshot attachRole = "snapshot"
)

type session struct {
	id        string
	name      string
	command   []string
	cwd       string
	createdAt time.Time

	cmd  *exec.Cmd
	ptmx *os.File
	cols int
	rows int

	mu          sync.Mutex
	scrollback  []byte
	done        chan struct{}
	exited      bool
	closeOnce   sync.Once
	waitOnce    sync.Once

	writeClaimed bool
	writerConn   *websocket.Conn
	observers    map[*websocket.Conn]struct{}
}

func (s *session) readLoop() {
	defer close(s.done)
	buf := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			data := buf[:n]
			s.mu.Lock()
			s.scrollback = append(s.scrollback, data...)
			if len(s.scrollback) > maxScrollback {
				s.scrollback = s.scrollback[len(s.scrollback)-maxScrollback:]
			}
			writer := s.writerConn
			observerSet := make([]*websocket.Conn, 0, len(s.observers))
			for conn := range s.observers {
				observerSet = append(observerSet, conn)
			}
			s.mu.Unlock()

			s.broadcastOutput(data, writer, observerSet)
		}
		if err != nil {
			s.markExited()
			s.appendExitMarker()
			s.wait()

			exitMsg := []byte("\r\n[Terminal exited]")
			s.mu.Lock()
			writer := s.writerConn
			observerSet := make([]*websocket.Conn, 0, len(s.observers))
			for conn := range s.observers {
				observerSet = append(observerSet, conn)
			}
			s.mu.Unlock()
			s.broadcastText(exitMsg, writer, observerSet)
			return
		}
	}
}

func (s *session) broadcastOutput(data []byte, writer *websocket.Conn, observers []*websocket.Conn) {
	if writer != nil {
		_ = writer.WriteMessage(websocket.BinaryMessage, data)
	}
	for _, conn := range observers {
		_ = conn.WriteMessage(websocket.BinaryMessage, data)
	}
}

func (s *session) broadcastText(data []byte, writer *websocket.Conn, observers []*websocket.Conn) {
	if writer != nil {
		_ = writer.WriteMessage(websocket.TextMessage, data)
	}
	for _, conn := range observers {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (s *session) snapshotInput() ([]byte, int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	scrollbackCopy := make([]byte, len(s.scrollback))
	copy(scrollbackCopy, s.scrollback)

	cols := s.cols
	rows := s.rows
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return scrollbackCopy, cols, rows
}

func (s *session) setSize(cols, rows int) {
	s.mu.Lock()
	s.cols = cols
	s.rows = rows
	s.mu.Unlock()
}

func (s *session) claimRole(attachMode string) attachRole {
	switch attachMode {
	case "observer":
		return roleObserver
	case "snapshot":
		return roleSnapshot
	case "screen", "interactive", "":
		s.mu.Lock()
		defer s.mu.Unlock()
		if !s.writeClaimed {
			s.writeClaimed = true
			return roleWriter
		}
		return roleObserver
	default:
		s.mu.Lock()
		defer s.mu.Unlock()
		if !s.writeClaimed {
			s.writeClaimed = true
			return roleWriter
		}
		return roleObserver
	}
}

func (s *session) registerConn(conn *websocket.Conn, role attachRole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch role {
	case roleWriter:
		s.writerConn = conn
	case roleObserver:
		if s.observers == nil {
			s.observers = make(map[*websocket.Conn]struct{})
		}
		s.observers[conn] = struct{}{}
	}
}

func (s *session) unregisterConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writerConn == conn {
		s.writerConn = nil
	}
	delete(s.observers, conn)
}

func (s *session) sendInitialFrame(conn *websocket.Conn, attachMode string) {
	scrollbackCopy, cols, rows := s.snapshotInput()
	if len(scrollbackCopy) == 0 {
		return
	}

	if attachMode == "screen" {
		if snapshot, ok := renderScreenSnapshot(scrollbackCopy, cols, rows); ok {
			conn.WriteMessage(websocket.BinaryMessage, snapshot)
			return
		}
	}

	altScreenActive := isAlternateScreenActive(scrollbackCopy)
	if !altScreenActive {
		conn.WriteMessage(websocket.BinaryMessage, []byte("\x1b[?1049l\x1b[0m"))
		scrollbackCopy = stripAlternateScreenPairs(scrollbackCopy)
		scrollbackCopy = stripTerminalQueries(scrollbackCopy)
	}
	conn.WriteMessage(websocket.BinaryMessage, scrollbackCopy)
}

func (s *session) sendRoleHandshake(conn *websocket.Conn, role attachRole) {
	payload, _ := json.Marshal(map[string]string{
		"type":        "attach_role",
		"attach_role": string(role),
	})
	_ = conn.WriteMessage(websocket.TextMessage, payload)
}

func (s *session) isWriter(conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writerConn == conn
}

func (s *session) status() string {
	if s.exited {
		return "exited"
	}
	return "running"
}

func (s *session) markExited() {
	s.mu.Lock()
	s.exited = true
	s.mu.Unlock()
}

func (s *session) appendExitMarker() {
	s.mu.Lock()
	defer s.mu.Unlock()

	exitMarker := []byte("\r\n[Terminal exited]")
	if bytes.HasSuffix(s.scrollback, exitMarker) {
		return
	}
	s.scrollback = append(s.scrollback, exitMarker...)
	if len(s.scrollback) > maxScrollback {
		s.scrollback = s.scrollback[len(s.scrollback)-maxScrollback:]
	}
}

func (s *session) wait() {
	s.waitOnce.Do(func() {
		if s.cmd != nil {
			s.cmd.Wait()
		}
	})
}

func (s *session) close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		if s.writerConn != nil {
			s.writerConn.Close()
			s.writerConn = nil
		}
		for conn := range s.observers {
			conn.Close()
		}
		s.observers = nil
		s.mu.Unlock()

		if s.ptmx != nil {
			s.ptmx.Close()
		}
		if s.cmd != nil && s.cmd.Process != nil && s.cmd.ProcessState == nil {
			s.cmd.Process.Kill()
		}
		s.wait()
	})
}

func (s *session) info(connected bool) SessionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	writerConnected := s.writerConn != nil
	observerCount := len(s.observers)
	if !connected {
		writerConnected = s.writerConn != nil
	}
	_ = observerCount
	return SessionInfo{
		ID:        s.id,
		Name:      s.name,
		Command:   append([]string(nil), s.command...),
		Cwd:       s.cwd,
		CreatedAt: s.createdAt,
		Status:    s.status(),
		Connected: connected || writerConnected || observerCount > 0,
	}
}

func (s *session) resize(cols, rows int) {
	if cols > 0 && rows > 0 {
		pty.Setsize(s.ptmx, &pty.Winsize{
			Rows: uint16(rows),
			Cols: uint16(cols),
		})
		s.setSize(cols, rows)
	}
}