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
	"github.com/hinshun/vt10x"
)

type attachRole string

const (
	roleWriter   attachRole = "writer"
	roleObserver attachRole = "observer"
	roleAttacher attachRole = "attacher"
	roleSnapshot attachRole = "snapshot"
)

type inputEventKind int

const (
	inputEventBytes inputEventKind = iota
	inputEventResize
)

type inputEvent struct {
	kind       inputEventKind
	data       []byte
	cols, rows int
}

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
	// screen is the persistent live VT (cell model). Source of truth for
	// attach_mode=snapshot / screen export. Updated on every PTY output chunk
	// and resized with the PTY; scrollback is secondary history only.
	screen vt10x.Terminal
	done   chan struct{}
	exited bool
	closeOnce sync.Once
	waitOnce  sync.Once

	writeClaimed bool
	writerConn   *websocket.Conn
	observers    map[*websocket.Conn]struct{}
	attachers    map[*websocket.Conn]struct{}

	inputCh   chan inputEvent
	inputOnce sync.Once
}

func (s *session) readLoop() {
	defer close(s.done)
	buf := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			data := buf[:n]
			s.mu.Lock()
			// Apply to live screen before (as) scrollback append so sticky cells
			// survive ring trim.
			if s.screen != nil {
				_, _ = s.screen.Write(data)
			}
			s.scrollback = append(s.scrollback, data...)
			if len(s.scrollback) > maxScrollback {
				s.scrollback = s.scrollback[len(s.scrollback)-maxScrollback:]
			}
			writer := s.writerConn
			observerSet := make([]*websocket.Conn, 0, len(s.observers))
			for conn := range s.observers {
				observerSet = append(observerSet, conn)
			}
			attacherSet := make([]*websocket.Conn, 0, len(s.attachers))
			for conn := range s.attachers {
				attacherSet = append(attacherSet, conn)
			}
			s.mu.Unlock()

			s.broadcastOutput(data, writer, observerSet, attacherSet)
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
			attacherSet := make([]*websocket.Conn, 0, len(s.attachers))
			for conn := range s.attachers {
				attacherSet = append(attacherSet, conn)
			}
			s.mu.Unlock()
			s.broadcastText(exitMsg, writer, observerSet, attacherSet)
			return
		}
	}
}

func (s *session) broadcastOutput(data []byte, writer *websocket.Conn, observers, attachers []*websocket.Conn) {
	if writer != nil {
		_ = writer.WriteMessage(websocket.BinaryMessage, data)
	}
	for _, conn := range observers {
		_ = conn.WriteMessage(websocket.BinaryMessage, data)
	}
	for _, conn := range attachers {
		_ = conn.WriteMessage(websocket.BinaryMessage, data)
	}
}

func (s *session) broadcastText(data []byte, writer *websocket.Conn, observers, attachers []*websocket.Conn) {
	if writer != nil {
		_ = writer.WriteMessage(websocket.TextMessage, data)
	}
	for _, conn := range observers {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
	for _, conn := range attachers {
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
	if s.screen != nil && cols > 0 && rows > 0 {
		s.screen.Resize(cols, rows)
	}
	s.mu.Unlock()
}

func (s *session) claimRole(attachMode string) attachRole {
	switch attachMode {
	case "observer":
		return roleObserver
	case "snapshot":
		return roleSnapshot
	case "attach":
		return roleAttacher
	case "screen", "interactive", "":
		s.mu.Lock()
		defer s.mu.Unlock()
		if !s.writeClaimed || s.writerConn == nil {
			s.writeClaimed = true
			return roleWriter
		}
		return roleObserver
	default:
		s.mu.Lock()
		defer s.mu.Unlock()
		if !s.writeClaimed || s.writerConn == nil {
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
	case roleAttacher:
		if s.attachers == nil {
			s.attachers = make(map[*websocket.Conn]struct{})
		}
		s.attachers[conn] = struct{}{}
	}
}

func (s *session) unregisterConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writerConn == conn {
		s.writerConn = nil
	}
	delete(s.observers, conn)
	delete(s.attachers, conn)
}

func (s *session) sendInitialFrame(conn *websocket.Conn, attachMode string) {
	// screen = interactive/writer path; snapshot = read-only one-shot (does not claim writer).
	// Both export the persistent live VT cells (source of truth), not a cold
	// replay of the truncated scrollback ring.
	if attachMode == "screen" || attachMode == "snapshot" {
		if snapshot, ok := s.exportLiveSnapshot(); ok {
			conn.WriteMessage(websocket.BinaryMessage, snapshot)
			return
		}
	}

	scrollbackCopy, cols, rows := s.snapshotInput()
	if len(scrollbackCopy) == 0 {
		return
	}

	// Fallback: cold replay if live screen is unavailable (should be rare).
	if attachMode == "screen" || attachMode == "snapshot" {
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

// exportLiveSnapshot walks the persistent live VT cells into the CUP frame
// format used by consumers. Holds session.mu so resize/output cannot race
// with the export walk.
func (s *session) exportLiveSnapshot() ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.screen == nil || len(s.scrollback) == 0 {
		return nil, false
	}
	cols := s.cols
	rows := s.rows
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return exportVTSnapshot(s.screen, cols, rows), true
}

func (s *session) sendRoleHandshake(conn *websocket.Conn, role attachRole) {
	_, cols, rows := s.snapshotInput()
	payload, _ := json.Marshal(map[string]any{
		"type":        "attach_role",
		"attach_role": string(role),
		"cols":        cols,
		"rows":        rows,
	})
	_ = conn.WriteMessage(websocket.TextMessage, payload)
}

func (s *session) isWriter(conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writerConn == conn
}

func (s *session) ensureInputLoop() {
	s.inputOnce.Do(func() {
		s.inputCh = make(chan inputEvent, 256)
		go s.inputLoop()
	})
}

func (s *session) enqueueBytes(data []byte) {
	s.ensureInputLoop()
	payload := append([]byte(nil), data...)
	s.inputCh <- inputEvent{kind: inputEventBytes, data: payload}
}

func (s *session) enqueueResize(cols, rows int) {
	s.ensureInputLoop()
	s.inputCh <- inputEvent{kind: inputEventResize, cols: cols, rows: rows}
}

func (s *session) inputLoop() {
	for event := range s.inputCh {
		switch event.kind {
		case inputEventBytes:
			_, _ = s.ptmx.Write(event.data)
		case inputEventResize:
			cols, rows := event.cols, event.rows
			for {
				select {
				case next, ok := <-s.inputCh:
					if !ok {
						s.resize(cols, rows)
						return
					}
					if next.kind == inputEventResize {
						cols, rows = next.cols, next.rows
						continue
					}
					s.resize(cols, rows)
					if next.kind == inputEventBytes {
						_, _ = s.ptmx.Write(next.data)
					}
					goto nextEvent
				default:
					s.resize(cols, rows)
					goto nextEvent
				}
			}
		}
	nextEvent:
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

// stopChild frees the OS PTY by closing the master and killing the child
// process, while leaving the session registered for metadata/scrollback.
// The session status becomes "exited" once readLoop (or this method) marks it.
// Safe to call more than once and concurrently with close().
func (s *session) stopChild() {
	if s.ptmx != nil {
		_ = s.ptmx.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil && s.cmd.ProcessState == nil {
		_ = s.cmd.Process.Kill()
	}
	// Reap the process so it does not linger as a zombie; waitOnce coordinates
	// with readLoop's wait() after EOF.
	s.wait()
	s.markExited()
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
		for conn := range s.attachers {
			conn.Close()
		}
		s.attachers = nil
		s.mu.Unlock()

		s.stopChild()
	})
}

func (s *session) info(connected bool) SessionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	writerConnected := s.writerConn != nil
	observerCount := len(s.observers)
	attacherCount := len(s.attachers)
	if !connected {
		writerConnected = s.writerConn != nil
	}
	return SessionInfo{
		ID:        s.id,
		Name:      s.name,
		Command:   append([]string(nil), s.command...),
		Cwd:       s.cwd,
		CreatedAt: s.createdAt,
		Status:    s.status(),
		Connected: connected || writerConnected || observerCount > 0 || attacherCount > 0,

		ObserverCount:   observerCount,
		AttacherCount:   attacherCount,
		WriterConnected: writerConnected,
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