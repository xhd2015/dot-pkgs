package ptywrap

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// writeFrame is called with outputMu held. Bound dead-browser writes so a
// half-open connection cannot indefinitely block a recovering attachment.
func writeFrame(conn *websocket.Conn, kind int, data []byte) error {
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := conn.WriteMessage(kind, data)
	if err != nil {
		// Unblock the read pump too; failed attachments must not accumulate
		// and delay every future replay behind repeated write timeouts.
		_ = conn.Close()
	}
	return err
}

// attachWeb atomically captures replay and subscribes to live output. The
// output lock keeps live frames after the handshake/replay; mu ensures no PTY
// output falls into the gap between the snapshot and registration.
func (s *session) attachWeb(conn *websocket.Conn) {
	s.outputMu.Lock()
	defer s.outputMu.Unlock()
	s.mu.Lock()
	replay := append([]byte(nil), s.scrollback...)
	cols, rows, status := s.cols, s.rows, s.status()
	if s.attachers == nil {
		s.attachers = make(map[*websocket.Conn]struct{})
	}
	s.attachers[conn] = struct{}{}
	s.mu.Unlock()

	role, _ := json.Marshal(map[string]any{
		"type": "attach_role", "attach_role": "attacher", "cols": cols, "rows": rows,
	})
	_ = writeFrame(conn, websocket.TextMessage, role)
	if len(replay) > 0 {
		if !isAlternateScreenActive(replay) {
			_ = writeFrame(conn, websocket.BinaryMessage, []byte("\x1b[?1049l\x1b[0m"))
			replay = stripTerminalQueries(stripAlternateScreenPairs(replay))
		}
		_ = writeFrame(conn, websocket.BinaryMessage, replay)
	}
	s.writeWebStateLocked(conn, status)
}

func (s *session) writeWebStateLocked(conn *websocket.Conn, status string) {
	payload, _ := json.Marshal(map[string]string{
		"type": "session_state", "session_id": s.id, "status": status,
	})
	_ = writeFrame(conn, websocket.TextMessage, payload)
}

func (s *session) sendWebState(conn *websocket.Conn, status string) {
	s.outputMu.Lock()
	defer s.outputMu.Unlock()
	s.writeWebStateLocked(conn, status)
}

func (s *session) sendWebPong(conn *websocket.Conn) {
	s.outputMu.Lock()
	defer s.outputMu.Unlock()
	_ = writeFrame(conn, websocket.TextMessage, []byte(`{"type":"pong"}`))
}
