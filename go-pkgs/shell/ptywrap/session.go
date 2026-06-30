package ptywrap

import (
	"bytes"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
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

	mu         sync.Mutex
	scrollback []byte
	conn       *websocket.Conn
	done       chan struct{}
	exited     bool
	closeOnce  sync.Once
	waitOnce   sync.Once
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
			ws := s.conn
			s.mu.Unlock()

			if ws != nil {
				ws.WriteMessage(websocket.BinaryMessage, data)
			}
		}
		if err != nil {
			s.markExited()
			s.appendExitMarker()
			s.wait()

			s.mu.Lock()
			ws := s.conn
			s.mu.Unlock()
			if ws != nil {
				ws.WriteMessage(websocket.TextMessage, []byte("\r\n[Terminal exited]"))
			}
			return
		}
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

func (s *session) attach(conn *websocket.Conn, attachMode string) {
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.conn = conn
	s.mu.Unlock()

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

func (s *session) detach(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == conn {
		s.conn = nil
	}
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
		if s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
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
	return SessionInfo{
		ID:        s.id,
		Name:      s.name,
		Command:   append([]string(nil), s.command...),
		Cwd:       s.cwd,
		CreatedAt: s.createdAt,
		Status:    s.status(),
		Connected: connected,
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